package wavstudio

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// Tabella internazionale ITU-R M.1677-1
var morseTable = map[rune]string{
	'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..", 'E': ".",
	'F': "..-.", 'G': "--.", 'H': "....", 'I': "..", 'J': ".---",
	'K': "-.-", 'L': ".-..", 'M': "--", 'N': "-.", 'O': "---",
	'P': ".--.", 'Q': "--.-", 'R': ".-.", 'S': "...", 'T': "-",
	'U': "..-", 'V': "...-", 'W': ".--", 'X': "-..-", 'Y': "-.--",
	'Z': "--..",

	'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-",
	'5': ".....", '6': "-....", '7': "--...", '8': "---..", '9': "----.",

	'.': ".-.-.-", ',': "--..--", '?': "..--..", '\'': ".----.",
	'!': "-.-.--", '/': "-..-.", '(': "-.--.", ')': "-.--.-",
	'&': ".-...", ':': "---...", ';': "-.-.-.", '=': "-...-",
	'+': ".-.-.", '-': "-....-", '_': "..--.-", '"': ".-..-.",
	'$': "...-..-", '@': ".--.-.",
}

func CreateWav(cfg *WavConfig) error {
	hasText := strings.TrimSpace(cfg.text) != ""
	hasFile := strings.TrimSpace(cfg.fname) != ""

	if hasText && hasFile {
		return fmt.Errorf("usa -text/argomento oppure -fname, non entrambi contemporaneamente")
	}
	if !hasText && !hasFile {
		return fmt.Errorf("specifica un testo tramite -text/argomento oppure un file tramite -fname")
	}
	if cfg.wpm <= 0 {
		return fmt.Errorf("il valore -wpm deve essere maggiore di 0")
	}
	if cfg.farnsworth < 0 || (cfg.farnsworth > 0 && cfg.farnsworth >= cfg.wpm) {
		return fmt.Errorf("-farnsworth deve essere maggiore di 0 e inferiore a -wpm")
	}
	if cfg.freq <= 0 || cfg.freq >= float64(cfg.sampleRate)/2 {
		return fmt.Errorf("frequenza del tono non valida rispetto al sample rate")
	}
	if cfg.sampleRate < 8000 {
		return fmt.Errorf("sample rate troppo basso (minimo 8000 Hz)")
	}
	if cfg.rampMs < 0 {
		return fmt.Errorf("la rampa non può essere negativa")
	}
	if cfg.repeat < 1 {
		return fmt.Errorf("-repeat deve essere almeno 1")
	}
	if cfg.pauseSec < 0 {
		return fmt.Errorf("-pause non può essere negativa")
	}
	if cfg.volume < 0 || cfg.volume > 1.0 {
		return fmt.Errorf("-volume deve essere compreso tra 0.0 e 1.0")
	}

	if hasFile {
		return processBatchFile(cfg)
	}

	return generateWAV(cfg, cfg.text, cfg.outFile)
}

func processBatchFile(cfg *WavConfig) error {
	f, err := os.Open(cfg.fname)
	if err != nil {
		return fmt.Errorf("impossibile aprire %s: %w", cfg.fname, err)
	}
	defer f.Close()

	if cfg.outDir != "" && cfg.outDir != "." {
		if err := os.MkdirAll(cfg.outDir, 0o755); err != nil {
			return fmt.Errorf("impossibile creare la cartella %s: %w", cfg.outDir, err)
		}
	}

	scanner := bufio.NewScanner(f)
	lineNum := 0
	generated := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		outPath := filepath.Join(cfg.outDir, sanitizeFilename(line)+".wav")
		if err := generateWAV(cfg, line, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "Riga %d (%q) saltata: %v\n", lineNum, line, err)
			continue
		}
		generated++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("errore durante la lettura di %s: %w", cfg.fname, err)
	}
	if generated == 0 {
		return fmt.Errorf("nessun file generato: il file è vuoto o non contiene righe valide")
	}

	return nil
}

func sanitizeFilename(line string) string {
	// Sostituisce gli spazi (anche multipli) con underscore
	name := strings.Join(strings.Fields(line), "_")

	// Neutralizza caratteri invalidi o problematici nei path (es. '/' o '?' presenti in Morse)
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	name = replacer.Replace(name)

	if name == "" || name == "." || name == ".." {
		return "morse_output"
	}
	return name
}

func generateWAV(cfg *WavConfig, text, outPath string) error {
	// Calcolo durate standard PARIS (50 unità per "PARIS ")
	ditSec := 1.2 / cfg.wpm
	dahSec := 3.0 * ditSec
	intraCharGapSec := ditSec

	// Unità di spaziatura (standard oppure Farnsworth ARRL)
	spaceUnitSec := ditSec
	if cfg.farnsworth > 0 {
		spaceUnitSec = (60.0*cfg.wpm - 37.2*cfg.farnsworth) / (19.0 * cfg.wpm * cfg.farnsworth)
	}
	charGapSec := 3.0 * spaceUnitSec
	wordGapSec := 7.0 * spaceUnitSec

	// Pre-codifica delle parole in sequenze di punti/linee
	rawWords := strings.Fields(strings.ToUpper(text))
	var encodedWords [][]string
	skipped := make(map[rune]bool)

	for _, w := range rawWords {
		var patterns []string
		for _, r := range w {
			code, ok := morseTable[r]
			if !ok {
				if !skipped[r] {
					fmt.Fprintf(os.Stderr, "Avviso (%q): carattere '%c' non presente nella tabella Morse, ignorato.\n", text, r)
					skipped[r] = true
				}
				continue
			}
			patterns = append(patterns, code)
		}
		if len(patterns) > 0 {
			encodedWords = append(encodedWords, patterns)
		}
	}

	if len(encodedWords) == 0 {
		return fmt.Errorf("nessun carattere valido da codificare")
	}

	var samples []int16
	rampSec := cfg.rampMs / 1000.0

	// 100 ms di silenzio iniziale per evitare tagli sui player smartphone
	samples = appendSilence(samples, 0.1, cfg.sampleRate)

	for rep := 0; rep < cfg.repeat; rep++ {
		if rep > 0 && cfg.pauseSec > 0 {
			samples = appendSilence(samples, cfg.pauseSec, cfg.sampleRate)
		}

		for wIdx, word := range encodedWords {
			if wIdx > 0 {
				samples = appendSilence(samples, wordGapSec, cfg.sampleRate)
			}

			for cIdx, pattern := range word {
				if cIdx > 0 {
					samples = appendSilence(samples, charGapSec, cfg.sampleRate)
				}

				for sIdx, sym := range pattern {
					if sIdx > 0 {
						samples = appendSilence(samples, intraCharGapSec, cfg.sampleRate)
					}

					dur := ditSec
					if sym == '-' {
						dur = dahSec
					}
					samples = appendTone(samples, dur, cfg.freq, rampSec, cfg.volume, cfg.sampleRate)
				}
			}
		}
	}

	// 100 ms di silenzio finale
	samples = appendSilence(samples, 0.1, cfg.sampleRate)

	if err := writeWAV(outPath, samples, cfg.sampleRate); err != nil {
		return err
	}

	totalDuration := float64(len(samples)) / float64(cfg.sampleRate)
	fmt.Printf("Generato: %s (durata: %.2fs, tono: %.0f Hz, velocità: %.0f WPM)\n",
		outPath, totalDuration, cfg.freq, cfg.wpm)

	if totalDuration > 30.0 {
		fmt.Fprintf(os.Stderr, "Nota (%s): la durata supera i 30 secondi (limite massimo per le suonerie su iPhone).\n", outPath)
	}

	return nil
}

func appendSilence(buf []int16, durationSec float64, sampleRate int) []int16 {
	n := int(math.Round(durationSec * float64(sampleRate)))
	if n <= 0 {
		return buf
	}
	return append(buf, make([]int16, n)...)
}

func appendTone(buf []int16, durationSec, freq, rampSec, volume float64, sampleRate int) []int16 {
	n := int(math.Round(durationSec * float64(sampleRate)))
	if n <= 0 {
		return buf
	}

	rampSamples := int(math.Round(rampSec * float64(sampleRate)))
	if 2*rampSamples > n {
		rampSamples = n / 2
	}

	sr := float64(sampleRate)
	maxAmp := volume * 32767.0

	for i := 0; i < n; i++ {
		env := 1.0
		if rampSamples > 0 {
			if i < rampSamples {
				// Salita a coseno rialzato (0 -> 1)
				env = 0.5 * (1.0 - math.Cos(math.Pi*float64(i)/float64(rampSamples)))
			} else if i >= n-rampSamples {
				// Discesa a coseno rialzato (1 -> 0)
				rem := n - 1 - i
				env = 0.5 * (1.0 - math.Cos(math.Pi*float64(rem)/float64(rampSamples)))
			}
		}

		t := float64(i) / sr
		sample := math.Sin(2.0*math.Pi*freq*t) * env * maxAmp

		if sample > 32767.0 {
			sample = 32767.0
		} else if sample < -32768.0 {
			sample = -32768.0
		}
		buf = append(buf, int16(math.Round(sample)))
	}

	return buf
}

func writeWAV(filename string, samples []int16, sampleRate int) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	numSamples := len(samples)
	dataBytes := numSamples * 2 // 16-bit mono = 2 byte per campione
	if dataBytes > math.MaxUint32-36 {
		return fmt.Errorf("file audio troppo grande per il formato WAV standard")
	}

	const (
		numChannels   uint16 = 1
		bitsPerSample uint16 = 16
	)
	byteRate := uint32(sampleRate) * uint32(numChannels) * uint32(bitsPerSample/8)
	blockAlign := numChannels * (bitsPerSample / 8)
	chunkSize := uint32(36 + dataBytes)
	subchunk2Size := uint32(dataBytes)

	// Header RIFF / WAVE (44 byte)
	if _, err := f.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, chunkSize); err != nil {
		return err
	}
	if _, err := f.Write([]byte("WAVE")); err != nil {
		return err
	}

	// Subchunk "fmt "
	if _, err := f.Write([]byte("fmt ")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil { // PCM header size
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(1)); err != nil { // AudioFormat 1 = PCM
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, numChannels); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, byteRate); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, blockAlign); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, bitsPerSample); err != nil {
		return err
	}

	// Subchunk "data"
	if _, err := f.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, subchunk2Size); err != nil {
		return err
	}

	// Serializzazione rapida dei campioni in little-endian
	rawPCM := make([]byte, dataBytes)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(rawPCM[i*2:], uint16(s))
	}

	_, err = f.Write(rawPCM)
	return err
}

