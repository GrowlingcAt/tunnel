import request from '../../request/axios';
export function deploy<T=any>(){
    const path = "/v1/app/deploy";
    return request.post<T>(path);
}