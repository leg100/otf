document.addEventListener('alpine:init', () => {
  Alpine.data('register', (hostname) => ({
    organization: '',
    get action() {
      if (this.organization === '') {
        return "https://" + hostname + "/settings/apps/new"
      }
      return "https://" + hostname + "/organizations/" + this.organization + "/settings/apps/new"
    },
  }))
})
