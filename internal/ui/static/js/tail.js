function isFunction(functionToCheck) {
  return functionToCheck && {}.toString.call(functionToCheck) === '[object Function]';
}

function debounce(func, wait) {
    var timeout;
    var waitFunc;

    return function() {
        if (isFunction(wait)) {
            waitFunc = wait;
        }
        else {
            waitFunc = function() { return wait };
        }

        var context = this, args = arguments;
        var later = function() {
            timeout = null;
            func.apply(context, args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, waitFunc());
    };
}

// reconnectFrequencySeconds doubles every retry
var reconnectFrequencySeconds = 1;

var reconnectFunc = debounce(function(path, phase, offset) {
    setupTail(path, phase, offset);
    // Double every attempt to avoid overwhelming server
    reconnectFrequencySeconds *= 2;
    // Max out at ~1 minute as a compromise between user experience and server load
    if (reconnectFrequencySeconds >= 64) {
        reconnectFrequencySeconds = 64;
    }
}, function() { return reconnectFrequencySeconds * 1000 });

function setupTail(path, phase, offset) {
    const url = `${path}?phase=${phase}&offset=${offset}`;
    var source = new EventSource(url);

    source.onopen = function(e) {
        // Reset reconnect frequency upon successful connection
        reconnectFrequencySeconds = 1;
    };

    source.onerror = function(e) {
        source.close();
        reconnectFunc(path, phase, offset);
    };

    source.addEventListener("log_update", (e) => {
        const obj = JSON.parse(e.data);

        // keep running tally of offset in case we need to reconnect
        offset = obj.offset;

        const elem = document.getElementById('tailed-' + phase + '-logs');
        const container = document.getElementById(phase + '-logs');

        // determine if user is following along at the end of the logs: either
        // within the logs' own scroll container, or, if the logs have been
        // expanded.
        const scrolls = container.scrollHeight > container.clientHeight;
        const atBottom = scrolls
            ? (container.scrollHeight - container.clientHeight - container.scrollTop) <= 100
            : (Math.floor(window.scrollY) + window.innerHeight) >= (document.body.scrollHeight - 100);

        elem.insertAdjacentHTML("beforeend", obj.html);

        // update toolbar button state.
        container.dispatchEvent(new CustomEvent("logs-appended"));

        if (!atBottom || container.clientHeight === 0) {
            // either the user has scrolled away from the end or phase is collapsed.
            return;
        }
        if (container.scrollHeight > container.clientHeight) {
            // scroll logs to reveal added content
            container.scrollTop = container.scrollHeight;
        } else {
            // scroll page to reveal added log content
            document.body.scrollIntoView(false);
        }
    });

    source.addEventListener("log_finished", (e) => {
        // no more logs to tail
        source.close();
    });
}

// phase_logs backs the toolbar sat above a phase's logs.
document.addEventListener('alpine:init', () => {
  Alpine.data('phase_logs', (hasLogs) => ({
    hasLogs: hasLogs,
    expanded: false,
    atTop: true,
    atBottom: true,
    init() {
      this.$nextTick(() => this.scrollToBottom());
    },
    // a collapsed phase has no height to scroll, wait until open
    // before showing the end of its logs.
    onToggle() {
      if (this.$el.open) {
        this.scrollToBottom();
      } else {
        this.refresh();
      }
    },
    // refresh top/bottom button state.
    refresh() {
      const el = this.$refs.logs;
      if (!el) {
        return;
      }
      // once expanded - or while the phase is collapsed there is nothing
      // to scroll, so both ends are considered reached.
      const scrollable = el.scrollHeight - el.clientHeight;
      this.atTop = scrollable <= 1 || el.scrollTop <= 0;
      this.atBottom = scrollable <= 1 || el.scrollTop >= scrollable - 1;
    },
    toggleExpanded() {
      this.expanded = !this.expanded;
      this.$nextTick(() => this.refresh());
    },
    scrollToTop() {
      this.$refs.logs.scrollTop = 0;
      this.refresh();
    },
    scrollToBottom() {
      const el = this.$refs.logs;
      el.scrollTop = el.scrollHeight;
      this.refresh();
    },
  }))
})
