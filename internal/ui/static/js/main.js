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

window.addEventListener("load", (e) => {
	document.body.addEventListener("htmx:afterSettle", function (evt) {
		Alpine.initTree(evt.detail.elt);
	});
});

function scrollMessageIntoView(message) {
	document.addEventListener("htmx:sseMessage", (e) => {
		// Only scroll if user has already scrolled to the bottom of the page.
		atBottom =
			Math.floor(window.scrollY) + window.innerHeight >=
			document.body.scrollHeight - 100;
		if (!atBottom) {
			return;
		}

		if (e.detail.type == message) {
			document.getElementById(e.detail.elt.id).scrollIntoView(true);
		}
	});
}
