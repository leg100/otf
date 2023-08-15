document.addEventListener('alpine:init', () => {
  Alpine.data('register', () => ({
    organization: '',
    get action() {
      if (this.organization === '') {
        return "https://github.com/settings/apps/new"
      }
      return "https://github.com/organizations/" + this.organization + "/settings/apps/new"
    },
  }))
})
