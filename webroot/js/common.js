new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    openId: '',
    poster: '',
    form: [],
    started: false,
    registered: false,
    sessions: [],
    userData: {},
    show: false
  },

  computed: {
    disable() {
      return !this.started || this.expired() || this.registered || this.isFull()
    },
    status() {
      if (this.expired()) return '报名已结束'
      if (this.registered) return '已报名'
      if (!this.started) return '报名尚未开始'
      if (this.isFull()) return '已报满'
      return '我要报名'
    }
  },

  mounted: async function () {
    try {
      let res = await axios.get(`/event-profile?event=${g_Event}&code=${g_WXCode}`)
      let data = res.data
      this.openId = data.openid
      this.poster = data.poster
      this.form = data.form
      this.show = true //show the submit button
      this.form.forEach((e) => {
        if (e.type == 'session') {
          this.userData.session = ''
        } else {
          this.userData[e.name] = ''
        }
      })

      await this.fechStatus()

      res = await axios.get(`/register-info?event=${g_Event}&openid=${this.openId}`)
      data = res.data
      if (data) {
        this.userData = data
        this.registered = true
      }
      setInterval(this.fechStatus, 1000)
    } catch(err) {
      console.log(err)
    }
  },

  methods: {
    async fechStatus() {
      try {
        const res = await axios.get(`/status?event=${g_Event}`)
        const data = res.data
        this.started = data.started
        this.sessions = data.sessions
        if (data.sessions.length == 1) {
          this.userData.session = 0
        }
      } catch(err) {
        console.log(err)
      }
    },

    isFull() {
      for (let i = 0; i < this.sessions.length; i++) {
        if (this.sessions[i].number < this.sessions[i].limit) {
          return false
        }
      }
      return true
    },

    expired() {
      for (let i = 0; i < this.sessions.length; i++) {
        if (!this.sessions[i].expired) {
            return false
        }
      }
      return true
    },

    formatAlertString(attr) {
      const form = this.form
      for (let i = 0; i < form.length; i++) {
        const component = form[i]
        if (attr == 'session' && component.type == 'session') {
          return '请选择' + component.name
        } else if (attr == component.name) {
          if (component.type == 'select') {
            return '请选择' + component.name 
          } 
          return '请输入' + component.name  
        }
      }
    },

    handleSubmit() {
      for (let attr in this.userData) {
        if (this.userData[attr] === '') {
          alert(this.formatAlertString(attr))
          return
        }
      }
      axios.post(`/submit-baoming?event=${g_Event}&openid=${this.openId}`, this.userData).then((response) => {
        const data = response.data
        if (data.errCode == 0) {
          this.registered = true
        }
        alert(data.errMsg)
      }).catch(function (error) {
        alert(error)
        console.log(error)
      })
    }
  }
})