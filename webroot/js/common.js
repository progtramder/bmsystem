function bmsystem(bmCallback) {
  new Vue({
    delimiters: ['${', '}'],
    el: '#app',
    data: {
      session: '',
      registered: false,
      started: false,
      expired: false,
      sessions: []
    },

    computed: {
      disable() {
        return !this.started || this.expired || this.registered || this.isFull()
      },
      status() {
        if (this.expired) return '报名已结束'
        if (this.registered) return '已报名'
        if (!this.started) return '报名尚未开始'
        if (this.isFull()) return '已报满'
        return '我要报名'
      }
    },

    created: function () {
      bmCallback.init(this)
    },

    mounted: function () {
      this.fechStatus()
      axios.get(`/register-info?event=${g_Event}&openid=${g_OpenId}`).then((response) => {
        const data = response.data
        if (data) {
          bmCallback.mount(this, data)
          this.session = data.session
          this.registered = true
        }
      }).catch(function (error) {
        console.log(error)
      })
      setInterval(this.fechStatus, 1000)
    },
    methods: {
      fechStatus() {
        axios.get(`/status?event=${g_Event}`).then((response) => {
          const data = response.data
          this.started = data.started
          this.expired = data.expired
          this.sessions = data.sessions
          if (data.sessions.length == 1) {
            this.session = 0
          }
        }).catch(function (error) {
          console.log(error)
        })
      },

      isFull() {
        for (let i = 0; i < this.sessions.length; i++) {
          if (this.sessions[i].number < this.sessions[i].limit) {
            return false
          }
        }
        return true
      },

      handleSubmit() {
        if (!bmCallback.validation(this)) {
          return
        }
        const postData = bmCallback.submit(this)
        axios.post(`/submit-baoming?event=${g_Event}&openid=${g_OpenId}&session=${this.session}`, postData).then((response) => {
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
}