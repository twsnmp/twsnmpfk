//@ts-ignore
import {fft,util} from 'fft-js';


export const doFFT = (signal:any, sampleRate:any) => {
  const np2 = 1 << (31 - Math.clz32(signal.length))
  while (signal.length !== np2) {
    signal.shift()
  }
  const phasors = fft(signal)
  const frequencies = util.fftFreq(phasors, sampleRate)
  const magnitudes = util.fftMag(phasors)
  const r = frequencies.map((f:any, ix:any) => {
    const p = f > 0.0 ? 1.0 / f : 0.0
    return { period: p, frequency: f, magnitude: magnitudes[ix] }
  })
  return r
}
