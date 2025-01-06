window.addEventListener("load", () => {
  const anchors = document.querySelectorAll("#sidemenu a");
  anchors.forEach((element) => {
    if (window.location.pathname == element.attributes.href.value) {
      element.classList.add("active");
    }
  });
});

// https://css-tricks.com/block-links-the-search-for-a-perfect-solution/#method-4-sprinkle-javascript-on-the-second-method
document.addEventListener("alpine:init", () => {
  Alpine.data("block_link", (block, link) => ({
    init() {
      block.classList.add("cursor-pointer", "hover:bg-gray-100");
      block.addEventListener("click", () => {
        isTextSelected = window.getSelection().toString();
        if (!isTextSelected) {
          location = link;
        }
      });
      const links = block.querySelectorAll("a");
      links.forEach((link) => {
        link.addEventListener("click", (e) => e.stopPropagation());
      });
    },
  }));
});
