new Vue({
    delimiters: ['${', '}'],
    el: '#app',
    data: {
        name: '',
        parent: '',
        phone: '',
        category: '',
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
                this.name = data['姓名']
                this.parent = data['身份']
                this.phone = data['手机号码']
                this.category = data['听众类型']
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
            const name = this.name
            const parent = this.parent
            const phone = this.phone
            const category = this.category
            const session = this.session
            if (name === '') {
                alert("请输入姓名")
                return
            }
            if (parent === '') {
                alert("请选择身份")
                return
            }
            if (!phone.match(/^\d{11}$/)) {
                alert("手机号码不正确")
                return
            }
            if (category === '') {
                alert("请选择听众类型")
                return
            }
            if (session === '') {
                alert("请选择场次")
                return
            }
            axios.post(`/submit-baoming?event=${g_Event}&openid=${g_OpenId}`, {
                '姓名': name,
                '身份': parent,
                '手机号码': phone,
                '听众类型': category
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
