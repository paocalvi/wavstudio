package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/paocalvi/wavstudio"
)

func main() {
	cfg := parseFlags()
	if err := wavstudio.CreateWav(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Errore: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *wavstudio.WavConfig {
	// Bind flags to local variables (fields are unexported; access via setters).
	var (
		text       string
		fname      string
		outFile    string
		outDir     string
		wpm        float64
		farnsworth float64
		freq       float64
		sampleRate int
		rampMs     float64
		repeat     int
		pauseSec   float64
		volume     float64
	)

	flag.StringVar(&text, "text", "", "Testo singolo da convertire in codice Morse (in alternativa come argomento finale)")
	flag.StringVar(&fname, "fname", "", "Percorso di un file di testo con una parola o frase per riga")
	flag.StringVar(&outFile, "out", "morse.wav", "Percorso del file .wav di output (usato in modalità testo singolo)")
	flag.StringVar(&outDir, "outdir", ".", "Cartella di destinazione per i file .wav generati con -fname")
	flag.Float64Var(&wpm, "wpm", 20.0, "Velocità dei caratteri in parole al minuto (standard PARIS)")
	flag.Float64Var(&farnsworth, "farnsworth", 0.0, "Velocità effettiva Farnsworth in WPM per le pause (0 = disabilitata)")
	flag.Float64Var(&freq, "freq", 800.0, "Frequenza del tono in Hz")
	flag.IntVar(&sampleRate, "rate", 44100, "Frequenza di campionamento in Hz")
	flag.Float64Var(&rampMs, "ramp", 5.0, "Durata della rampa di salita/discesa in millisecondi (anti-click)")
	flag.IntVar(&repeat, "repeat", 1, "Numero di ripetizioni del messaggio")
	flag.Float64Var(&pauseSec, "pause", 1.5, "Pausa in secondi tra una ripetizione e l'altra")
	flag.Float64Var(&volume, "volume", 0.8, "Volume del segnale (da 0.0 a 1.0)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso:\n  %s [opzioni] [testo]\n  %s [opzioni] -fname lista.txt\n\nOpzioni:\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Trailing positional argument as text fallback.
	if text == "" && flag.NArg() > 0 {
		text = strings.Join(flag.Args(), " ")
	}

	// Build config: technical params via constructor, content/path params via setters.
	cfg := wavstudio.NewWavConfig(wpm, farnsworth, freq, sampleRate, rampMs, repeat, pauseSec, volume)
	cfg.SetText(text)
	cfg.SetFname(fname)
	cfg.SetOutFile(outFile)
	cfg.SetOutDir(outDir)

	return cfg
}
