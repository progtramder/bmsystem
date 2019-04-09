bmsystem({
    init(vue) {
        vue.student  = ''
        vue.gender   = ''
        vue.idType   = ''
        vue.idNumber = ''
    },
    mount(vue, data) {
        vue.student  = data['孩子姓名']
        vue.gender   = data['性别']
        vue.idType   = data['证件类型']
        vue.idNumber = data['证件号码']
    },
    validation(vue) {
        const student  = vue.student
        const gender   = vue.gender
        const idType   = vue.idType
        const idNumber = vue.idNumber
        const session  = vue.session
        if (student === '') {
            alert("请输入孩子姓名")
            return false
        }
        if (gender === '') {
            alert("请选择孩子性别")
            return false
        }
        if (idType === '') {
            alert("请选择证件类型")
            return false
        }
        if (idNumber === '') {
            alert("请输入证件号码")
            return false
        }
        if (session === '') {
            alert("请选择体验场次")
            return false
        }
        return true
    },
    submit(vue) {
        return {
            '孩子姓名': vue.student,
            '性别':    vue.gender,
            '证件类型': vue.idType,
            '证件号码': vue.idNumber
        }
    }
 })