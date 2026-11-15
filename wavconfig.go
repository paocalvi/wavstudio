package wavstudio

// WavConfig holds all the parameters needed to generate a Morse-code WAV file.
type WavConfig struct {
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
}

// NewWavConfig creates a WavConfig pre-populated with the technical audio
// parameters. Text / file / output-path fields are left at their zero values
// and should be set afterwards via SetText, SetFname, SetOutFile, SetOutDir.
//
// Parameters (in order):
//
//	wpm        – Sending speed in words per minute (standard PARIS)
//	farnsworth – Farnsworth spacing speed in WPM (0 = disabled)
//	freq       – Tone frequency in Hz
//	sampleRate – Audio sample rate in Hz
//	rampMs     – Attack/release ramp duration in milliseconds
//	repeat     – Number of times the message is repeated (≥ 1)
//	pauseSec   – Pause between repetitions in seconds
//	volume     – Output volume, 0.0 – 1.0
func NewWavConfig(wpm, farnsworth, freq float64, sampleRate int, rampMs float64, repeat int, pauseSec, volume float64) *WavConfig {
	return &WavConfig{
		wpm:        wpm,
		farnsworth: farnsworth,
		freq:       freq,
		sampleRate: sampleRate,
		rampMs:     rampMs,
		repeat:     repeat,
		pauseSec:   pauseSec,
		volume:     volume,
	}
}

// --- Getters ---

// GetText returns the text to encode as Morse code.
func (c *WavConfig) GetText() string { return c.text }

// GetFname returns the path of the batch input file.
func (c *WavConfig) GetFname() string { return c.fname }

// GetOutFile returns the output WAV file path.
func (c *WavConfig) GetOutFile() string { return c.outFile }

// GetOutDir returns the output directory for batch mode.
func (c *WavConfig) GetOutDir() string { return c.outDir }

// GetWpm returns the sending speed in words per minute.
func (c *WavConfig) GetWpm() float64 { return c.wpm }

// GetFarnsworth returns the Farnsworth spacing speed (0 = disabled).
func (c *WavConfig) GetFarnsworth() float64 { return c.farnsworth }

// GetFreq returns the tone frequency in Hz.
func (c *WavConfig) GetFreq() float64 { return c.freq }

// GetSampleRate returns the audio sample rate in Hz.
func (c *WavConfig) GetSampleRate() int { return c.sampleRate }

// GetRampMs returns the attack/release ramp duration in milliseconds.
func (c *WavConfig) GetRampMs() float64 { return c.rampMs }

// GetRepeat returns the number of times the message is repeated.
func (c *WavConfig) GetRepeat() int { return c.repeat }

// GetPauseSec returns the pause between repetitions in seconds.
func (c *WavConfig) GetPauseSec() float64 { return c.pauseSec }

// GetVolume returns the output volume (0.0 – 1.0).
func (c *WavConfig) GetVolume() float64 { return c.volume }

// --- Setters ---

// SetText sets the text to encode as Morse code.
func (c *WavConfig) SetText(v string) { c.text = v }

// SetFname sets the path of the batch input file.
func (c *WavConfig) SetFname(v string) { c.fname = v }

// SetOutFile sets the output WAV file path.
func (c *WavConfig) SetOutFile(v string) { c.outFile = v }

// SetOutDir sets the output directory for batch mode.
func (c *WavConfig) SetOutDir(v string) { c.outDir = v }

// SetWpm sets the sending speed in words per minute.
func (c *WavConfig) SetWpm(v float64) { c.wpm = v }

// SetFarnsworth sets the Farnsworth spacing speed (0 = disabled).
func (c *WavConfig) SetFarnsworth(v float64) { c.farnsworth = v }

// SetFreq sets the tone frequency in Hz.
func (c *WavConfig) SetFreq(v float64) { c.freq = v }

// SetSampleRate sets the audio sample rate in Hz.
func (c *WavConfig) SetSampleRate(v int) { c.sampleRate = v }

// SetRampMs sets the attack/release ramp duration in milliseconds.
func (c *WavConfig) SetRampMs(v float64) { c.rampMs = v }

// SetRepeat sets the number of times the message is repeated.
func (c *WavConfig) SetRepeat(v int) { c.repeat = v }

// SetPauseSec sets the pause between repetitions in seconds.
func (c *WavConfig) SetPauseSec(v float64) { c.pauseSec = v }

// SetVolume sets the output volume (0.0 – 1.0).
func (c *WavConfig) SetVolume(v float64) { c.volume = v }

