new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
      student: '',
      gender: '',
      idType: '',
      idNumber: '',
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

  mounted: function() {
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

      axios.get(`/register-info?event=${g_Event}&openid=${g_OpenId}`).then((response) => {
          const data = response.data
          if (data) {
              this.student = data['孩子姓名']
              this.gender = data['性别']
              this.idType = data['证件类型']
              this.idNumber = data['证件号码']
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
          const student = this.student
          const gender = this.gender
          const idType = this.idType
          const idNumber = this.idNumber
          const session = this.session
          if (student === '') {
              alert("请输入孩子姓名")
              return
          }
          if (gender === '') {
              alert("请选择孩子性别")
              return
          }
          if (idType === '') {
            alert("请选择证件类型")
            return
          }
          if (idNumber === '') {
              alert("请输入证件号码")
              return
          }
          if (session === '') {
              alert("请选择体验场次")
              return
          }
          axios.post(`/submit-baoming?event=${g_Event}&openid=${g_OpenId}&session=${session}`, {
              '孩子姓名': student,
              '性别': gender,
              '证件类型': idType,
              '证件号码': idNumber
          }).then((response) => {
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
