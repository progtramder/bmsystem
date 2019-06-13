new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
      school: '',
      event: '',
      started: false,
      sessions: [] 
  },

  computed: {
    expired: function() {
      for (let i = 0; i < this.sessions.length; i++) {
        if (!this.sessions[i].expired) {
            return false
        }
      }
      return true
    }
  },

  mounted: function() {
      this.school = this.getQueryString("school")
      this.event = this.getQueryString("event")
      const title = document.getElementsByTagName("title")[0]
      title.innerText = this.school
      this.fechStatus()
      setInterval(this.fechStatus, 1000)
  },
  methods: {
      getQueryString(name) {
          const reg = new RegExp('(^|&)' + name + '=([^&]*)(&|$)', 'i')
          const query = window.location.search.substr(1).match(reg)
          if (query != null) {
              return unescape(query[2])
          }
          return '';
      },

      fechStatus() {
          axios.get(`/status?school=${this.school}&event=${this.event}`).then((response) => {
              const data = response.data
              this.started = data.started
              this.sessions = data.sessions
          }).catch(function (error) {
              console.log(error)
          })
      },
      
      handleSubmit() {
          axios.get(`/start-baoming?school=${this.school}&event=${this.event}`).then((response) => {
              this.started = true
          }).catch(function (error) {
              console.log(error)
          })
      },

      handleDetail() {
        window.location.href = `/report/${this.event}.xlsx`
      }
  }
})
