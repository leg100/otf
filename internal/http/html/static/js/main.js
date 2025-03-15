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

// https://css-tricks.com/block-links-the-search-for-a-perfect-solution/#method-4-sprinkle-javascript-on-the-second-method
document.addEventListener("alpine:init", () => {
	Alpine.data("block_link", (block, link) => ({
		init() {
			block.classList.add("cursor-pointer", "hover:bg-base-200");
			block.addEventListener("click", (e) => {
				console.info(e);
				isTextSelected = window.getSelection().toString();
				if (!isTextSelected) {
					location = link;
				}
			});
			links = block.querySelectorAll("a, button");
			links.forEach((link) => {
				link.addEventListener("click", (e) => e.stopPropagation());
			});
		},
	}));
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
