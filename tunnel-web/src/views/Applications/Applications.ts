import request from  '../../request/axios';

export function getApplications<T=any>() {
    const path = "/v1/app";
    return request.get<T>(path)
}

export interface app {
    id: number,
    name: string,
    type: string,
    local_ip: string,
    local_port: number,
    entry_domain: string,
    entry_port: number,
    proxy:string,
    proxy_port: number,
}
export function addApplication<T=any>(params: app) {
    const path = "/v1/app";
    let form = new FormData();
    form.append("name", params.name);
    form.append("type", params.type);
    form.append("local_ip", params.local_ip);
    form.append("local_port", params.local_port.toString());
    return request.post<T>(path, form);
}

export function editApplication<T=any>(params: app) {
    const path = "/v1/app";
    let form = new FormData();
    form.append("app_id", params.id.toString());
    form.append("name", params.name);
    form.append("type", params.type);
    form.append("local_ip", params.local_ip);
    form.append("local_port", params.local_port.toString());
    return request.put<T>(path, form);
}