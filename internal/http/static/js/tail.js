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

        // determine if user has scrolled to the very bottom of page
        atBottom = (Math.floor(window.scrollY) + window.innerHeight) >= (document.body.scrollHeight - 100);

        const elem = document.getElementById('tailed-' + phase + '-logs');
        elem.insertAdjacentHTML("beforeend", obj.html);

        if (atBottom) {
            // scroll page down to reveal added log content
            document.body.scrollIntoView(false);
        }
    });

    source.addEventListener("log_finished", (e) => {
        // no more logs to tail
        source.close();
    });
}
