document.addEventListener('alpine:init', () => {
  Alpine.data('run_elapsed_time', (created, elapsed) => ({
    init() {
        setInterval(() => {
          elapsed = Date.now() - created;
        }, 500);
    },
    elapsed: elapsed,
    formatDuration(ms) {
      if (ms < 1000) {
        return `${ms}ms`
      }
      if (ms < 60000) {
        s = ms / 1000;
        return `${s}s`
      }
      m = ms / 60000;
      s = ms % 60000;
      return `${m}m${s}s`
    },
  }))
})
