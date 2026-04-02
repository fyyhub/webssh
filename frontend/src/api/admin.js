import request from '@/utils/request'

export function getAdminConfig() {
    return request.get('/admin/config')
}

export function verifyAdminPassword(password) {
    return request.post('/admin/verify', { password })
}
