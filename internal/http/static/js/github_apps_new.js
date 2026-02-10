document.addEventListener('alpine:init', () => {
  Alpine.data('action', (hostname, manifest) => ({
    organization: '',
    manifest: JSON.parse(manifest),
    public: false,
    get action() {
      if (this.organization === '') {
        return "https://" + hostname + "/settings/apps/new"
      }
      return "https://" + hostname + "/organizations/" + this.organization + "/settings/apps/new"
    },
  }))
})
