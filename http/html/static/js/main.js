window.addEventListener('load', (e) => {
    // enables copy to clipboard functionality for identifiers
    const identifiers = document.querySelectorAll('.clipboard-copyable');
    identifiers.forEach(function(id) {
        id.addEventListener('click', function(event) {
            content = event.target.innerHTML;
            navigator.clipboard.writeText(content);
            notification = event.target.parentNode.querySelector('.copied-notification');
            if (notification === null) {
                let span = document.createElement('span');
                span.className = 'copied-notification';
                span.innerHTML = 'copied!';
                // show notification momentarily
                event.target.parentNode.appendChild(span);
                setTimeout(function() {
                    event.target.parentNode.removeChild(span);
                }, 1000);
            }
        });
    });
});

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
//
// reconnectFrequencySeconds doubles every retry
var reconnectFrequencySeconds = 1;

var reconnectFunc = debounce(function(path, phase, offset, stream) {
    setupTail(path, phase, offset, stream);
    // Double every attempt to avoid overwhelming server
    reconnectFrequencySeconds *= 2;
    // Max out at ~1 minute as a compromise between user experience and server load
    if (reconnectFrequencySeconds >= 64) {
        reconnectFrequencySeconds = 64;
    }
}, function() { return reconnectFrequencySeconds * 1000 });

function setupTail(path, phase, offset, stream) {
    const url = `${path}?phase=${phase}&offset=${offset}&stream=${stream}`;
    var source = new EventSource(url);

    source.onopen = function(e) {
        // Reset reconnect frequency upon successful connection
        reconnectFrequencySeconds = 1;
    };

    source.onerror = function(e) {
        source.close();
        reconnectFunc(path, phase, offset, stream);
    };

    source.addEventListener("new-log-chunk", (e) => {
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

    source.addEventListener("finished", (e) => {
        // no more logs to tail
        source.close();
    });
}

function watchRunUpdates(path, stream, run) {
    const url = `${path}?stream=${stream}&run-id=${run}`;
    var source = new EventSource(url);

    source.addEventListener("run-status-update", (e) => {
        const obj = JSON.parse(e.data);

        const runStatus = document.getElementById(obj.id + '-status-');
        elem.outerHTML = obj['run-status']

        const planStatus = document.getElementById('plan-status');
        elem.outerHTML = obj['plan-status']

        const applyStatus = document.getElementById('apply-status');
        elem.outerHTML = obj['apply-status']
    });
}

function watchWorkspaceUpdates(path, stream) {
    const url = `${path}?stream=${stream}`;
    var source = new EventSource(url);

    source.addEventListener("run-latest-update", (e) => {
        const obj = JSON.parse(e.data);

        const latestRunElem = document.getElementById('latest-run');
        latestRunElem.outerHTML = obj['html']
    });
}

function watchRuns(path, stream) {
    const url = `${path}?stream=${stream}`;
    var source = new EventSource(url);

    const listElem = document.getElementById('content-list');

    source.addEventListener('run_created', (e) => {
        const obj = JSON.parse(e.data);

        listElem.insertAdjacentHTML("afterbegin", obj.html);
    });

    source.addEventListener('run_status_update', (e) => {
        const obj = JSON.parse(e.data);

        const runElem = document.getElementById(obj.id);
        runElem.remove();
        listElem.insertAdjacentHTML("afterbegin", obj.html);
    });
}
