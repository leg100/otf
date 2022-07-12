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
