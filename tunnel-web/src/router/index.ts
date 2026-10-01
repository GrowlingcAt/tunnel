import { createMemoryHistory, createRouter } from 'vue-router'
import internal from '../views/internal/internal.vue'
import Deploy from '../views/Deploy/Deploy.vue'
import Applications from '../views/Applications/Applications.vue'
const routes = [
    {
        path: '/',
        component: internal,
        children: [
            {path:'apps/',component:Applications},
            {path:'deploy/',component:Deploy},
        ]
    }
]

const router =  createRouter({
    history: createMemoryHistory(),
    routes,
})

export default router