window.addEventListener('load', (e) => {
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
