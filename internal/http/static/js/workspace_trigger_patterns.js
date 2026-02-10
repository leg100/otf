document.addEventListener('alpine:init', () => {
  Alpine.data('workspace_trigger_patterns', (paths = []) => ({
    paths: paths || [],
    newPath: '',
    get addPattern() {
      if (this.newPath === '') return
      if (this.paths?.includes(this.newPath)) return
      this.paths.push(this.newPath)
      this.newPath = ''
    },
    deletePattern(pattern) {
      this.paths = this.paths.filter(i => i !== pattern)
    },
  }))
})
