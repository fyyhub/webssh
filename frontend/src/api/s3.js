import request from '@/utils/request'

export function getS3Config() {
    return request.get('/s3/config')
}

export function listS3Objects(prefix) {
    return request.get(`/s3/list?prefix=${encodeURIComponent(prefix || '')}`)
}
