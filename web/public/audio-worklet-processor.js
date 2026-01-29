/**
 * AudioWorklet processor for real-time audio analysis and visualization
 * Runs in a separate audio thread for low-latency processing
 */
class AudioAnalyzerProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
    this.bufferSize = 2048;
    this.buffer = new Float32Array(this.bufferSize);
    this.bufferIndex = 0;
    this.lastRMS = 0;
    this.lastPeak = 0;

    // Handle messages from main thread
    this.port.onmessage = (event) => {
      if (event.data.type === 'reset') {
        this.bufferIndex = 0;
        this.lastRMS = 0;
        this.lastPeak = 0;
      }
    };
  }

  process(inputs, outputs, parameters) {
    const input = inputs[0];
    const output = outputs[0];

    if (input.length === 0) {
      return true;
    }

    const inputChannel = input[0];
    const outputChannel = output[0];

    // Pass through audio
    if (inputChannel && outputChannel) {
      outputChannel.set(inputChannel);
    }

    // Analyze audio in buffer
    if (inputChannel) {
      for (let i = 0; i < inputChannel.length; i++) {
        this.buffer[this.bufferIndex] = inputChannel[i];
        this.bufferIndex = (this.bufferIndex + 1) % this.bufferSize;
      }

      // Calculate RMS and peak every 128 samples
      if (this.bufferIndex % 128 === 0) {
        let sumSquares = 0;
        let peak = 0;

        for (let i = 0; i < this.bufferSize; i++) {
          const sample = this.buffer[i];
          sumSquares += sample * sample;
          if (Math.abs(sample) > peak) {
            peak = Math.abs(sample);
          }
        }

        const rms = Math.sqrt(sumSquares / this.bufferSize);

        // Smooth values
        this.lastRMS = this.lastRMS * 0.9 + rms * 0.1;
        this.lastPeak = Math.max(this.lastPeak * 0.95, peak);

        // Send analysis data to main thread
        this.port.postMessage({
          type: 'analysis',
          rms: this.lastRMS,
          peak: this.lastPeak,
          buffer: this.buffer.slice(0, 512) // Send partial buffer for waveform
        });
      }
    }

    return true;
  }
}

/**
 * Stream decoder processor for progressive audio loading
 * Decodes incoming audio chunks and outputs PCM samples
 */
class StreamDecoderProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
    this.sampleQueue = [];
    this.isPlaying = false;
    this.fadeIn = true;
    this.fadeOut = false;
    this.fadeLength = 128; // Samples for fade
    this.fadeIndex = 0;

    this.port.onmessage = (event) => {
      switch (event.data.type) {
        case 'samples':
          // Receive decoded samples from main thread
          this.sampleQueue.push(...event.data.samples);
          break;
        case 'play':
          this.isPlaying = true;
          this.fadeIn = true;
          this.fadeIndex = 0;
          break;
        case 'pause':
          this.fadeOut = true;
          this.fadeIndex = 0;
          break;
        case 'stop':
          this.isPlaying = false;
          this.sampleQueue = [];
          break;
        case 'seek':
          // Clear buffer on seek
          this.sampleQueue = [];
          break;
      }
    };
  }

  process(inputs, outputs, parameters) {
    const output = outputs[0];

    if (!this.isPlaying || output.length === 0) {
      // Output silence
      for (let channel = 0; channel < output.length; channel++) {
        output[channel].fill(0);
      }
      return true;
    }

    const outputChannel = output[0];
    const samplesToOutput = outputChannel.length;

    for (let i = 0; i < samplesToOutput; i++) {
      let sample = 0;

      if (this.sampleQueue.length > 0) {
        sample = this.sampleQueue.shift();
      }

      // Apply fade in
      if (this.fadeIn) {
        const fade = this.fadeIndex / this.fadeLength;
        sample *= fade;
        this.fadeIndex++;
        if (this.fadeIndex >= this.fadeLength) {
          this.fadeIn = false;
        }
      }

      // Apply fade out
      if (this.fadeOut) {
        const fade = 1 - (this.fadeIndex / this.fadeLength);
        sample *= Math.max(0, fade);
        this.fadeIndex++;
        if (this.fadeIndex >= this.fadeLength) {
          this.fadeOut = false;
          this.isPlaying = false;
        }
      }

      outputChannel[i] = sample;
    }

    // Copy to other channels
    for (let channel = 1; channel < output.length; channel++) {
      output[channel].set(outputChannel);
    }

    // Report buffer status
    this.port.postMessage({
      type: 'bufferStatus',
      buffered: this.sampleQueue.length,
      isPlaying: this.isPlaying
    });

    return true;
  }
}

// Register both processors
registerProcessor('audio-analyzer-processor', AudioAnalyzerProcessor);
registerProcessor('stream-decoder-processor', StreamDecoderProcessor);
