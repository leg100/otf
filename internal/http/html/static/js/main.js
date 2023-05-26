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
