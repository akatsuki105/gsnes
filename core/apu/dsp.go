package apu

type DSPChannel struct {
	// pitch
	pitch           uint16
	pitchCounter    uint16
	pitchModulation bool
	// brr decoding
	decodeBuffer  [19]int16 // 16 samples per brr-block, +3 for interpolation
	srcn          byte
	decodeOffset  uint16
	previousFlags byte // from last sample
	old           int16
	older         int16
	useNoise      bool
	// adsr, envelope, gain
	adsrRates    [4]uint16 // attack, decay, sustain, gain
	rateCounter  uint16
	adsrState    byte // 0: attack, 1: decay, 2: sustain, 3: gain, 4: release
	sustainLevel uint16
	useGain      bool
	gainMode     byte
	directGain   bool
	gainValue    uint16 // for direct gain
	gain         uint16
	// keyon/off
	keyOn  bool
	keyOff bool
	// output
	sampleOut  int16 // final sample, to be multiplied by channel volume
	volumeL    int8
	volumeR    int8
	echoEnable bool
}

type DSP struct {
	apu *apu

	// mirror ram
	ram [0x80]byte
	// 8 channels
	channel [8]DSPChannel
	// overarching
	dirPage       uint16
	evenCycle     bool
	mute          bool
	reset         bool
	masterVolumeL int8
	masterVolumeR int8
	// noise
	noiseSample  int16
	noiseRate    uint16
	noiseCounter uint16
	// echo
	echoWrites      bool
	echoVolumeL     int8
	echoVolumeR     int8
	feedbackVolume  int8
	echoBufferAddr  uint16
	echoDelay       uint16
	echoRemain      uint16
	echoBufferIndex uint16
	firBufferIndex  uint8
	firValues       [8]int8
	firBufferL      [8]int16
	firBufferR      [8]int16
	// sample buffer (1 frame at 32040 Hz: 534 samples, *2 for stereo)
	sampleBuffer [534 * 2]int16
	sampleOffset uint16 // current offset in samplebuffer
}

var rateValues [32]int = [32]int{
	0, 2048, 1536, 1280, 1024, 768, 640, 512,
	384, 320, 256, 192, 160, 128, 96, 80,
	64, 48, 40, 32, 24, 20, 16, 12,
	10, 8, 6, 5, 4, 3, 2, 1,
}

var gaussValues [512]int = [512]int{
	0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000,
	0x001, 0x001, 0x001, 0x001, 0x001, 0x001, 0x001, 0x001, 0x001, 0x001, 0x001, 0x002, 0x002, 0x002, 0x002, 0x002,
	0x002, 0x002, 0x003, 0x003, 0x003, 0x003, 0x003, 0x004, 0x004, 0x004, 0x004, 0x004, 0x005, 0x005, 0x005, 0x005,
	0x006, 0x006, 0x006, 0x006, 0x007, 0x007, 0x007, 0x008, 0x008, 0x008, 0x009, 0x009, 0x009, 0x00A, 0x00A, 0x00A,
	0x00B, 0x00B, 0x00B, 0x00C, 0x00C, 0x00D, 0x00D, 0x00E, 0x00E, 0x00F, 0x00F, 0x00F, 0x010, 0x010, 0x011, 0x011,
	0x012, 0x013, 0x013, 0x014, 0x014, 0x015, 0x015, 0x016, 0x017, 0x017, 0x018, 0x018, 0x019, 0x01A, 0x01B, 0x01B,
	0x01C, 0x01D, 0x01D, 0x01E, 0x01F, 0x020, 0x020, 0x021, 0x022, 0x023, 0x024, 0x024, 0x025, 0x026, 0x027, 0x028,
	0x029, 0x02A, 0x02B, 0x02C, 0x02D, 0x02E, 0x02F, 0x030, 0x031, 0x032, 0x033, 0x034, 0x035, 0x036, 0x037, 0x038,
	0x03A, 0x03B, 0x03C, 0x03D, 0x03E, 0x040, 0x041, 0x042, 0x043, 0x045, 0x046, 0x047, 0x049, 0x04A, 0x04C, 0x04D,
	0x04E, 0x050, 0x051, 0x053, 0x054, 0x056, 0x057, 0x059, 0x05A, 0x05C, 0x05E, 0x05F, 0x061, 0x063, 0x064, 0x066,
	0x068, 0x06A, 0x06B, 0x06D, 0x06F, 0x071, 0x073, 0x075, 0x076, 0x078, 0x07A, 0x07C, 0x07E, 0x080, 0x082, 0x084,
	0x086, 0x089, 0x08B, 0x08D, 0x08F, 0x091, 0x093, 0x096, 0x098, 0x09A, 0x09C, 0x09F, 0x0A1, 0x0A3, 0x0A6, 0x0A8,
	0x0AB, 0x0AD, 0x0AF, 0x0B2, 0x0B4, 0x0B7, 0x0BA, 0x0BC, 0x0BF, 0x0C1, 0x0C4, 0x0C7, 0x0C9, 0x0CC, 0x0CF, 0x0D2,
	0x0D4, 0x0D7, 0x0DA, 0x0DD, 0x0E0, 0x0E3, 0x0E6, 0x0E9, 0x0EC, 0x0EF, 0x0F2, 0x0F5, 0x0F8, 0x0FB, 0x0FE, 0x101,
	0x104, 0x107, 0x10B, 0x10E, 0x111, 0x114, 0x118, 0x11B, 0x11E, 0x122, 0x125, 0x129, 0x12C, 0x130, 0x133, 0x137,
	0x13A, 0x13E, 0x141, 0x145, 0x148, 0x14C, 0x150, 0x153, 0x157, 0x15B, 0x15F, 0x162, 0x166, 0x16A, 0x16E, 0x172,
	0x176, 0x17A, 0x17D, 0x181, 0x185, 0x189, 0x18D, 0x191, 0x195, 0x19A, 0x19E, 0x1A2, 0x1A6, 0x1AA, 0x1AE, 0x1B2,
	0x1B7, 0x1BB, 0x1BF, 0x1C3, 0x1C8, 0x1CC, 0x1D0, 0x1D5, 0x1D9, 0x1DD, 0x1E2, 0x1E6, 0x1EB, 0x1EF, 0x1F3, 0x1F8,
	0x1FC, 0x201, 0x205, 0x20A, 0x20F, 0x213, 0x218, 0x21C, 0x221, 0x226, 0x22A, 0x22F, 0x233, 0x238, 0x23D, 0x241,
	0x246, 0x24B, 0x250, 0x254, 0x259, 0x25E, 0x263, 0x267, 0x26C, 0x271, 0x276, 0x27B, 0x280, 0x284, 0x289, 0x28E,
	0x293, 0x298, 0x29D, 0x2A2, 0x2A6, 0x2AB, 0x2B0, 0x2B5, 0x2BA, 0x2BF, 0x2C4, 0x2C9, 0x2CE, 0x2D3, 0x2D8, 0x2DC,
	0x2E1, 0x2E6, 0x2EB, 0x2F0, 0x2F5, 0x2FA, 0x2FF, 0x304, 0x309, 0x30E, 0x313, 0x318, 0x31D, 0x322, 0x326, 0x32B,
	0x330, 0x335, 0x33A, 0x33F, 0x344, 0x349, 0x34E, 0x353, 0x357, 0x35C, 0x361, 0x366, 0x36B, 0x370, 0x374, 0x379,
	0x37E, 0x383, 0x388, 0x38C, 0x391, 0x396, 0x39B, 0x39F, 0x3A4, 0x3A9, 0x3AD, 0x3B2, 0x3B7, 0x3BB, 0x3C0, 0x3C5,
	0x3C9, 0x3CE, 0x3D2, 0x3D7, 0x3DC, 0x3E0, 0x3E5, 0x3E9, 0x3ED, 0x3F2, 0x3F6, 0x3FB, 0x3FF, 0x403, 0x408, 0x40C,
	0x410, 0x415, 0x419, 0x41D, 0x421, 0x425, 0x42A, 0x42E, 0x432, 0x436, 0x43A, 0x43E, 0x442, 0x446, 0x44A, 0x44E,
	0x452, 0x455, 0x459, 0x45D, 0x461, 0x465, 0x468, 0x46C, 0x470, 0x473, 0x477, 0x47A, 0x47E, 0x481, 0x485, 0x488,
	0x48C, 0x48F, 0x492, 0x496, 0x499, 0x49C, 0x49F, 0x4A2, 0x4A6, 0x4A9, 0x4AC, 0x4AF, 0x4B2, 0x4B5, 0x4B7, 0x4BA,
	0x4BD, 0x4C0, 0x4C3, 0x4C5, 0x4C8, 0x4CB, 0x4CD, 0x4D0, 0x4D2, 0x4D5, 0x4D7, 0x4D9, 0x4DC, 0x4DE, 0x4E0, 0x4E3,
	0x4E5, 0x4E7, 0x4E9, 0x4EB, 0x4ED, 0x4EF, 0x4F1, 0x4F3, 0x4F5, 0x4F6, 0x4F8, 0x4FA, 0x4FB, 0x4FD, 0x4FF, 0x500,
	0x502, 0x503, 0x504, 0x506, 0x507, 0x508, 0x50A, 0x50B, 0x50C, 0x50D, 0x50E, 0x50F, 0x510, 0x511, 0x511, 0x512,
	0x513, 0x514, 0x514, 0x515, 0x516, 0x516, 0x517, 0x517, 0x517, 0x518, 0x518, 0x518, 0x518, 0x518, 0x519, 0x519,
}

func NewDSP(apu *apu) *DSP {
	return &DSP{
		apu: apu,
	}
}

func (dsp *DSP) Reset() {
	dsp.ram[0x7c] = 0xff // set ENDx
	dsp.mute = true
	dsp.reset = true
	dsp.noiseSample = -0x4000
	dsp.echoDelay = 1
	dsp.echoRemain = 1
}

func (dsp *DSP) Cycle() {
	var totalL int = 0
	var totalR int = 0
	for i := 0; i < len(dsp.channel); i++ {
		dsp.cycleChannel(i)
		totalL += (int(dsp.channel[i].sampleOut) * int(dsp.channel[i].volumeL)) >> 6
		totalR += (int(dsp.channel[i].sampleOut) * int(dsp.channel[i].volumeR)) >> 6
		totalL = clamp16bit(totalL)
		totalR = clamp16bit(totalR)
	}
	totalL = (totalL * int(dsp.masterVolumeL)) >> 7
	totalR = (totalR * int(dsp.masterVolumeR)) >> 7
	totalL = clamp16bit(totalL)
	totalR = clamp16bit(totalR)
	dsp.handleEcho(&totalL, &totalR)
	if dsp.mute {
		totalL = 0
		totalR = 0
	}
	dsp.handleNoise()
	// put it in the samplebuffer
	dsp.sampleBuffer[dsp.sampleOffset*2] = int16(totalL)
	dsp.sampleBuffer[dsp.sampleOffset*2+1] = int16(totalR)

	// prevent sampleOffset from going above 534-1 (out of sampleBuffer bounds)
	if dsp.sampleOffset < 533 {
		dsp.sampleOffset++
	}

	dsp.evenCycle = !dsp.evenCycle
}

func (dsp *DSP) handleEcho(outputL *int, outputR *int) {
	// get value out of ram
	var addr uint16 = dsp.echoBufferAddr + dsp.echoBufferIndex*4
	dsp.firBufferL[dsp.firBufferIndex] = (int16(dsp.apu.ram[addr]) + (int16(dsp.apu.ram[(addr+1)&0xffff]) << 8))
	dsp.firBufferL[dsp.firBufferIndex] >>= 1
	dsp.firBufferR[dsp.firBufferIndex] = (int16(dsp.apu.ram[(addr+2)&0xffff]) + (int16(dsp.apu.ram[(addr+3)&0xffff]) << 8))
	dsp.firBufferR[dsp.firBufferIndex] >>= 1
	// calculate FIR-sum
	var sumL, sumR int = 0, 0
	for i := 0; i < len(dsp.channel); i++ {
		sumL += (int(dsp.firBufferL[(int(dsp.firBufferIndex)+i+1)&0x7]) * int(dsp.firValues[i])) >> 6
		sumR += (int(dsp.firBufferR[(int(dsp.firBufferIndex)+i+1)&0x7]) * int(dsp.firValues[i])) >> 6
		if i == 6 {
			// clip to 16-bit before last addition
			sumL = int(int16(sumL & 0xffff)) // clip 16-bit
			sumR = int(int16(sumR & 0xffff)) // clip 16-bit
		}
	}
	sumL = clamp16bit(sumL)
	sumR = clamp16bit(sumR)
	// modify output with sum
	var outL int = *outputL + ((sumL * int(dsp.echoVolumeL)) >> 7)
	var outR int = *outputR + ((sumR * int(dsp.echoVolumeR)) >> 7)
	*outputL = clamp16bit(outL)
	*outputR = clamp16bit(outR)
	// get echo input
	var inL, inR int = 0, 0
	for i := 0; i < len(dsp.channel); i++ {
		if dsp.channel[i].echoEnable {
			inL += (int(dsp.channel[i].sampleOut) * int(dsp.channel[i].volumeL)) >> 6
			inR += (int(dsp.channel[i].sampleOut) * int(dsp.channel[i].volumeR)) >> 6
			inL = clamp16bit(inL)
			inR = clamp16bit(inR)
		}
	}
	// write this to ram
	inL += (sumL * int(dsp.feedbackVolume)) >> 7
	inR += (sumR * int(dsp.feedbackVolume)) >> 7
	inL = clamp16bit(inL)
	inR = clamp16bit(inR)
	inL &= 0xfffe
	inR &= 0xfffe
	if dsp.echoWrites {
		dsp.apu.ram[addr] = byte(inL & 0xff)
		dsp.apu.ram[(addr+1)&0xffff] = byte(inL >> 8)
		dsp.apu.ram[(addr+2)&0xffff] = byte(inR & 0xff)
		dsp.apu.ram[(addr+3)&0xffff] = byte(inR >> 8)
	}
	// handle indexes
	dsp.firBufferIndex++
	dsp.firBufferIndex &= 7
	dsp.echoBufferIndex++
	dsp.echoRemain--
	if dsp.echoRemain == 0 {
		dsp.echoRemain = dsp.echoDelay
		dsp.echoBufferIndex = 0
	}
}

func (dsp *DSP) cycleChannel(ch int) {
	// handle pitch counter
	var pitch uint16 = dsp.channel[ch].pitch
	if ch > 0 && dsp.channel[ch].pitchModulation {
		var factor int = (int(dsp.channel[ch-1].sampleOut) >> 4) + 0x400
		pitch = uint16((int(pitch) * factor) >> 10)
		if pitch > 0x3fff {
			pitch = 0x3fff
		}
	}
	var newCounter int = int(dsp.channel[ch].pitchCounter) + int(pitch)
	if newCounter > 0xffff {
		// next sample
		dsp.decodeBrr(ch)
	}
	dsp.channel[ch].pitchCounter = uint16(newCounter)
	var sample int16 = 0
	if dsp.channel[ch].useNoise {
		sample = dsp.noiseSample
	} else {
		sample = dsp.getSample(ch, int(dsp.channel[ch].pitchCounter)>>12, (int(dsp.channel[ch].pitchCounter)>>4)&0xff)
	}
	if dsp.evenCycle {
		// handle keyon/off (every other cycle)
		if dsp.channel[ch].keyOff {
			// go to release
			dsp.channel[ch].adsrState = 4
		} else if dsp.channel[ch].keyOn {
			dsp.channel[ch].keyOn = false
			// restart current sample
			dsp.channel[ch].previousFlags = 0
			var samplePointer uint16 = dsp.dirPage + (4 * uint16(dsp.channel[ch].srcn))
			dsp.channel[ch].decodeOffset = uint16(dsp.apu.ram[samplePointer])
			dsp.channel[ch].decodeOffset |= uint16(dsp.apu.ram[(samplePointer+1)&0xffff]) << 8
			// zero clear
			for i := 0; i < len(dsp.channel[ch].decodeBuffer); i++ {
				dsp.channel[ch].decodeBuffer[i] = 0
			}
			dsp.channel[ch].gain = 0
			if dsp.channel[ch].useGain {
				dsp.channel[ch].adsrState = 3
			} else {
				dsp.channel[ch].adsrState = 0
			}
		}
	}
	// handle reset
	if dsp.reset {
		dsp.channel[ch].adsrState = 4
		dsp.channel[ch].gain = 0
	}
	// handle envelope/adsr
	var doingDirectGain bool = dsp.channel[ch].adsrState != 4 && dsp.channel[ch].useGain && dsp.channel[ch].directGain
	var rate uint16
	if dsp.channel[ch].adsrState == 4 {
		rate = 0
	} else {
		rate = dsp.channel[ch].adsrRates[dsp.channel[ch].adsrState]
	}
	if dsp.channel[ch].adsrState != 4 && !doingDirectGain && rate != 0 {
		dsp.channel[ch].rateCounter++
	}
	if dsp.channel[ch].adsrState == 4 || (!doingDirectGain && dsp.channel[ch].rateCounter >= rate && rate != 0) {
		if dsp.channel[ch].adsrState != 4 {
			dsp.channel[ch].rateCounter = 0
		}
		dsp.handleGain(ch)
	}
	if doingDirectGain {
		dsp.channel[ch].gain = dsp.channel[ch].gainValue
	}
	// set outputs
	dsp.ram[(ch<<4)|8] = byte(dsp.channel[ch].gain >> 4)
	// WARNING: convert to unsigned int
	sample = int16(uint16((int(sample) * int(dsp.channel[ch].gain)) >> 11))
	dsp.ram[(ch<<4)|9] = byte(sample >> 7)
	dsp.channel[ch].sampleOut = sample
}

func (dsp *DSP) handleGain(ch int) {
	switch dsp.channel[ch].adsrState {
	case 0: // attack
		var rate uint16 = dsp.channel[ch].adsrRates[dsp.channel[ch].adsrState]
		if rate == 1 {
			dsp.channel[ch].gain += 1024
		} else {
			dsp.channel[ch].gain += 32
		}
		if dsp.channel[ch].gain >= 0x7e0 {
			dsp.channel[ch].adsrState = 1
		}
		if dsp.channel[ch].gain > 0x7ff {
			dsp.channel[ch].gain = 0x7ff
		}
	case 1:
		// decay
		dsp.channel[ch].gain -= ((dsp.channel[ch].gain - 1) >> 8) + 1
		if dsp.channel[ch].gain < dsp.channel[ch].sustainLevel {
			dsp.channel[ch].adsrState = 2
		}
	case 2:
		// sustain
		sub := ((dsp.channel[ch].gain - 1) >> 8) + 1
		// WARNING: overflow handling
		if dsp.channel[ch].gain >= sub {
			dsp.channel[ch].gain -= ((dsp.channel[ch].gain - 1) >> 8) + 1
		} else {
			dsp.channel[ch].gain = 0
		}
	case 3:
		// gain
		switch dsp.channel[ch].gainMode {
		case 0:
			// linear decrease
			dsp.channel[ch].gain -= 32
			// decreasing below 0 will underflow to above 0x7ff
			if dsp.channel[ch].gain > 0x7ff {
				dsp.channel[ch].gain = 0
			}
		case 1:
			// exponential decrease
			dsp.channel[ch].gain -= ((dsp.channel[ch].gain - 1) >> 8) + 1
		case 2:
			// linear increase
			dsp.channel[ch].gain += 32
			if dsp.channel[ch].gain > 0x7ff {
				dsp.channel[ch].gain = 0x7ff
			}
		case 3:
			// bent increase
			if dsp.channel[ch].gain < 0x600 {
				dsp.channel[ch].gain += 32
			} else {
				dsp.channel[ch].gain += 8
			}
			if dsp.channel[ch].gain > 0x7ff {
				dsp.channel[ch].gain = 0x7ff
			}
		}
	case 4:
		// release
		dsp.channel[ch].gain -= 8
		// decreasing below 0 will underflow to above 0x7ff
		if dsp.channel[ch].gain > 0x7ff {
			dsp.channel[ch].gain = 0
		}
	}
}

func (dsp *DSP) getSample(ch int, sampleNum int, offset int) int16 {
	var news int16 = dsp.channel[ch].decodeBuffer[sampleNum+3]
	var olds int16 = dsp.channel[ch].decodeBuffer[sampleNum+2]
	var olders int16 = dsp.channel[ch].decodeBuffer[sampleNum+1]
	var oldests int16 = dsp.channel[ch].decodeBuffer[sampleNum]
	var out int = (gaussValues[0xff-offset] * int(oldests)) >> 10

	out += (gaussValues[0x1ff-offset] * int(olders)) >> 10
	out += (gaussValues[0x100+offset] * int(olds)) >> 10
	out = int(int16(out & 0xffff)) // clip 16-bit
	out += (gaussValues[offset] * int(news)) >> 10
	out = clamp16bit(out) // clamp 16-bit

	return int16(out >> 1)
}

func (dsp *DSP) decodeBrr(ch int) {
	// copy last 3 samples (16-18) to first 3 for interpolation
	dsp.channel[ch].decodeBuffer[0] = dsp.channel[ch].decodeBuffer[16]
	dsp.channel[ch].decodeBuffer[1] = dsp.channel[ch].decodeBuffer[17]
	dsp.channel[ch].decodeBuffer[2] = dsp.channel[ch].decodeBuffer[18]
	// handle flags from previous block
	if dsp.channel[ch].previousFlags == 1 || dsp.channel[ch].previousFlags == 3 {
		// loop sample
		var samplePointer uint16 = uint16(int(dsp.dirPage) + (4 * int(dsp.channel[ch].srcn)))
		dsp.channel[ch].decodeOffset = uint16(dsp.apu.ram[(samplePointer+2)&0xffff])
		dsp.channel[ch].decodeOffset |= uint16(dsp.apu.ram[(samplePointer+3)&0xffff]) << 8
		if dsp.channel[ch].previousFlags == 1 {
			// also release and clear gain
			dsp.channel[ch].adsrState = 4
			dsp.channel[ch].gain = 0
		}
		dsp.ram[0x7c] |= 1 << ch // set ENDx
	}
	var header byte = dsp.apu.ram[dsp.channel[ch].decodeOffset]
	dsp.channel[ch].decodeOffset++
	var shift int = int(header) >> 4
	var filter int = int(header&0xc) >> 2
	dsp.channel[ch].previousFlags = header & 0x3
	var curByte byte = 0
	var old int = int(dsp.channel[ch].old)
	var older int = int(dsp.channel[ch].older)
	for i := 0; i < 16; i++ {
		var s int = 0
		if (i & 1) > 0 {
			s = int(curByte) & 0x0f
		} else {
			curByte = dsp.apu.ram[dsp.channel[ch].decodeOffset]
			dsp.channel[ch].decodeOffset++
			s = int(curByte) >> 4
		}
		if s > 7 {
			s -= 16
		}
		if shift <= 0xc {
			s = (s << shift) >> 1
		} else {
			s = (s >> 3) << 12
		}
		switch filter {
		case 1:
			s += old + (-old >> 4)
			break
		case 2:
			s += 2*old + ((3 * -old) >> 5) - older + (older >> 4)
			break
		case 3:
			s += 2*old + ((13 * -old) >> 6) - older + ((3 * older) >> 4)
			break
		}
		s = clamp16bit(s)
		s = int((int16((s & 0x7fff) << 1)) >> 1) // clip 15-bit
		older = old
		old = s
		dsp.channel[ch].decodeBuffer[i+3] = int16(s)
	}
	dsp.channel[ch].older = int16(older)
	dsp.channel[ch].old = int16(old)
}

func (dsp *DSP) handleNoise() {
	if dsp.noiseRate != 0 {
		dsp.noiseCounter++
	}
	if dsp.noiseCounter >= dsp.noiseRate && dsp.noiseRate != 0 {
		var bit int = (int(dsp.noiseSample) & 1) ^ ((int(dsp.noiseSample) >> 1) & 1)
		dsp.noiseSample = ((dsp.noiseSample >> 1) & 0x3fff) | (int16(bit) << 14)
		dsp.noiseSample = (int16((dsp.noiseSample & 0x7fff) << 1)) >> 1
		dsp.noiseCounter = 0
	}
}

func (dsp *DSP) Read(addr byte) byte {
	return dsp.ram[addr]
}

func (dsp *DSP) Write(addr byte, value byte) {
	var ch int = int(addr) >> 4
	switch addr {
	case 0x00, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70:
		dsp.channel[ch].volumeL = int8(value)
	case 0x01, 0x11, 0x21, 0x31, 0x41, 0x51, 0x61, 0x71:
		dsp.channel[ch].volumeR = int8(value)
	case 0x02, 0x12, 0x22, 0x32, 0x42, 0x52, 0x62, 0x72:
		dsp.channel[ch].pitch = (dsp.channel[ch].pitch & 0x3f00) | uint16(value)
	case 0x03, 0x13, 0x23, 0x33, 0x43, 0x53, 0x63, 0x73:
		dsp.channel[ch].pitch = ((dsp.channel[ch].pitch & 0x00ff) | (uint16(value) << 8)) & 0x3fff
	case 0x04, 0x14, 0x24, 0x34, 0x44, 0x54, 0x64, 0x74:
		dsp.channel[ch].srcn = value
	case 0x05, 0x15, 0x25, 0x35, 0x45, 0x55, 0x65, 0x75:
		dsp.channel[ch].adsrRates[0] = uint16(rateValues[(value&0xf)*2+1])
		dsp.channel[ch].adsrRates[1] = uint16(rateValues[((value&0x70)>>4)*2+16])
		dsp.channel[ch].useGain = (value & 0x80) == 0
	case 0x06, 0x16, 0x26, 0x36, 0x46, 0x56, 0x66, 0x76:
		dsp.channel[ch].adsrRates[2] = uint16(rateValues[value&0x1f])
		dsp.channel[ch].sustainLevel = (((uint16(value) & 0xe0) >> 5) + 1) * 0x100
	case 0x07, 0x17, 0x27, 0x37, 0x47, 0x57, 0x67, 0x77:
		dsp.channel[ch].directGain = (value & 0x80) == 0
		if (value & 0x80) > 0 {
			dsp.channel[ch].gainMode = (value & 0x60) >> 5
			dsp.channel[ch].adsrRates[3] = uint16(rateValues[value&0x1f])
		} else {
			dsp.channel[ch].gainValue = (uint16(value) & 0x7f) * 16
		}
	case 0x0c:
		dsp.masterVolumeL = int8(value)
	case 0x1c:
		dsp.masterVolumeR = int8(value)
	case 0x2c:
		dsp.echoVolumeL = int8(value)
	case 0x3c:
		dsp.echoVolumeR = int8(value)
	case 0x4c:
		for i := 0; i < len(dsp.channel); i++ {
			dsp.channel[i].keyOn = (value & (1 << i)) > 0
		}
	case 0x5c:
		for i := 0; i < len(dsp.channel); i++ {
			dsp.channel[i].keyOff = (value & (1 << i)) > 0
		}
	case 0x6c:
		dsp.reset = (value & 0x80) > 0
		dsp.mute = (value & 0x40) > 0
		dsp.echoWrites = (value & 0x20) == 0
		dsp.noiseRate = uint16(rateValues[value&0x1f])
	case 0x7c:
		value = 0 // any write clears ENDx
	case 0x0d:
		dsp.feedbackVolume = int8(value)
	case 0x2d:
		for i := 0; i < len(dsp.channel); i++ {
			dsp.channel[i].pitchModulation = (value & (1 << i)) > 0
		}
	case 0x3d:
		for i := 0; i < len(dsp.channel); i++ {
			dsp.channel[i].useNoise = (value & (1 << i)) > 0
		}
	case 0x4d:
		for i := 0; i < len(dsp.channel); i++ {
			dsp.channel[i].echoEnable = (value & (1 << i)) > 0
		}
	case 0x5d:
		dsp.dirPage = uint16(value) << 8
	case 0x6d:
		dsp.echoBufferAddr = uint16(value) << 8
	case 0x7d:
		dsp.echoDelay = (uint16(value) & 0xf) * 512 // 2048-byte steps, stereo sample is 4 bytes
		if dsp.echoDelay == 0 {
			dsp.echoDelay = 1
		}
	case 0x0f, 0x1f, 0x2f, 0x3f, 0x4f, 0x5f, 0x6f, 0x7f:
		dsp.firValues[ch] = int8(value)
	}

	dsp.ram[addr] = value
}

// utilities

func clamp16bit(total int) int {
	// clamp 16-bit
	if total < -0x8000 {
		return -0x8000
	} else {
		if total > 0x7fff {
			return 0x7fff
		}
	}

	return total
}

func (dsp *DSP) getSamples(sampleData []int16, samplesPerFrame int) {
	// resample from 534 samples per frame to wanted value
	var adder float64 = 534.0 / float64(samplesPerFrame)
	var location float64 = 0.0
	for i := 0; i < samplesPerFrame; i++ {
		sampleData[i*2] = dsp.sampleBuffer[int(location)*2]
		sampleData[i*2+1] = dsp.sampleBuffer[int(location)*2+1]
		location += adder
	}
	dsp.sampleOffset = 0
}
