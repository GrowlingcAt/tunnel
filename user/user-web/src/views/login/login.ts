import {get} from '../../request/request'

export interface sysParams{
    sys:string,
}
export interface officialCallbackParams extends sysParams {
    ticket:string,
}

export function loginMethods<T=any>(params:sysParams) {
    const path = "/v1/login/methods"
    return get<T>({url:path,data:{"sys":params.sys}})
}

export function loginOfficialCallback<T=any>(params:officialCallbackParams){
    const path = "/v1/login/official/callback"
    return get<T>({url:path,data:{"sys":params.sys,"ticket":params.ticket}})
}