new Vue({
  el: '#app',
  data: {
      password: ''
  },

  methods: {
    handleLogin() {
      var exdate = new Date()
      exdate.setDate(exdate.getDate() + 365)
      document.cookie = "admin_password=" + md5(escape(this.password)) + ";expires=" + exdate.toGMTString()
      window.location.href = "/develop"
    }
  }
})
