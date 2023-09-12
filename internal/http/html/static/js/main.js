// https://daverupert.com/2017/11/happier-html5-forms/
window.addEventListener('load', (e) => {
  const inputs = document.querySelectorAll("input, select, textarea");
  inputs.forEach(input => {
    input.addEventListener(
      "invalid",
      event => {
        input.classList.add("error");
      },
      false
    );
    input.addEventListener("blur", function() {
      input.checkValidity();
    });
  });
});

// window.addEventListener('htmx:afterSwap', (e) => {
//   const widgets = e.detail.elt.querySelectorAll(".widget");
//   widgets.forEach(widget => {
//     blockLink(widget);
//   });
// });

// https://css-tricks.com/block-links-the-search-for-a-perfect-solution/#method-4-sprinkle-javascript-on-the-second-method
document.addEventListener('alpine:init', () => {
  Alpine.data('block_link', (block, link) => ({
    init() {
      block.classList.add("cursor-pointer", "hover:bg-gray-100");
      block.addEventListener("click", (e) => {
        isTextSelected = window.getSelection().toString();
        if (!isTextSelected) {
          location = link;
        }
      });
      links = block.querySelectorAll("a");
      links.forEach(link => {
        link.addEventListener("click", (e) => e.stopPropagation());
      });
    }
  }))
})
