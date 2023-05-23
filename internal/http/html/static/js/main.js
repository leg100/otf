window.addEventListener('load', (e) => {
    // enables copy to clipboard functionality
    const identifiers = document.querySelectorAll('.clipboard-icon');
    identifiers.forEach(function(id) {
        id.addEventListener('click', function(event) {
            content = event.target.previousElementSibling.innerHTML;
            navigator.clipboard.writeText(content);

            // show notification momentarily but only if there isn't already a notification showing
            notification = event.target.parentNode.querySelector('.copied-notification');
            if (notification === null) {
                let span = document.createElement('span');
                span.className = 'copied-notification';
                span.innerHTML = 'copied!';
                event.target.parentNode.appendChild(span);
                setTimeout(function() {
                    event.target.parentNode.removeChild(span);
                }, 1000);
            }
        });
    });

    // https://daverupert.com/2017/11/happier-html5-forms/
    const inputs = document.querySelectorAll("input, select, textarea");
    inputs.forEach(input => {
      input.addEventListener(
        "invalid",
        event => {
          input.classList.add("error");
        },
        false
      );
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

function watchRunUpdates(path, stream, run) {
    const url = `${path}?stream=${stream}&run-id=${run}`;
    var source = new EventSource(url);

    source.addEventListener("updated", (e) => {
        const obj = JSON.parse(e.data);

        const runItem = document.getElementById(obj.id);
        runItem.outerHTML = obj['run-item-html'];

        const planStatus = document.getElementById('plan-status');
        planStatus.outerHTML = obj['plan-status-html'];

        const applyStatus = document.getElementById('apply-status');
        applyStatus.outerHTML = obj['apply-status-html'];

        const runActions = document.getElementById('run-actions-container');
        runActions.innerHTML = obj['run-actions-html'];

        // if user is at/near very bottom of page then scroll down to
        // bring any new content beneath the viewport into view.
        atBottom = (Math.floor(window.scrollY) + window.innerHeight) >= (document.body.scrollHeight - 100);
        if (atBottom) {
            document.body.scrollIntoView(false);
        }
    });
}

function watchWorkspaceUpdates(path) {
    var source = new EventSource(path);

    source.addEventListener("updated", (e) => {
        const obj = JSON.parse(e.data);

        const latestRunElem = document.getElementById('latest-run');
        latestRunElem.innerHTML = obj['run-item-html']
    });
}

function watchRuns(path) {
    var source = new EventSource(path);

    const listElem = document.getElementById('content-list');

    source.addEventListener('created', (e) => {
        const obj = JSON.parse(e.data);

        listElem.insertAdjacentHTML("afterbegin", obj['run-item-html']);
    });

    source.addEventListener('updated', (e) => {
        const obj = JSON.parse(e.data);

        const runElem = document.getElementById(obj.id);
        runElem.remove();
        listElem.insertAdjacentHTML("afterbegin", obj['run-item-html']);
    });
}
