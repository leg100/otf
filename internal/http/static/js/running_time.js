document.addEventListener('alpine:init', () => {
  Alpine.data('running_time', (created, elapsed, done) => ({
    created: created,
    elapsed: elapsed,
    done: done,
    init() {
      if (this.done) {
        return
      }
      setInterval(() => {
        this.elapsed = Date.now() - this.created;
      }, 500);
    },
    formatDuration(ms) {
      if (ms < 1000) {
        return `${ms}ms`
      }
      if (ms < 60000) {
        s = Math.floor(ms / 1000);
        return `${s}s`
      }
      m = Math.floor(ms / 60000);
      s = Math.floor((ms % 60000) / 1000);
      return `${m}m${s}s`
    },
  }))
})
