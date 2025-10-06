// https://daverupert.com/2017/11/happier-html5-forms/
window.addEventListener("load", (e) => {
	const inputs = document.querySelectorAll("input, select, textarea");
	inputs.forEach((input) => {
		input.addEventListener(
			"invalid",
			(event) => {
				input.classList.add("error");
			},
			false,
		);
		input.addEventListener("blur", function () {
			input.checkValidity();
		});
	});
});

window.addEventListener("htmx:wsConfigSend", function (evt) {
	const msg = evt.detail.parameters;

	// remove headers from message before sending because we have no use for them.
	delete msg.HEADERS;

	// don't send JSON, but send url-encoded query instead
	let query = new URLSearchParams();
	Object.entries(msg).forEach(([k, v]) => {
		if (Array.isArray(v)) {
			v.forEach((vv) => query.append(k, vv));
		} else query.append(k, v);
	});
	const params = query.toString();
	evt.detail.messageBody = params;
});
