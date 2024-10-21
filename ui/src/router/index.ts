import { createRouter, createWebHistory } from 'vue-router'
import Login from '../components/Login.vue'
import Profile from '../components/Profile.vue'
import OrganizationList from '../components/OrganizationList.vue'
import OrganizationItem from '../components/OrganizationItem.vue'
import WorkspaceList from '../components/WorkspaceList.vue'
import AdminLogin from '../components/AdminLogin.vue'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            redirect: () => {
                return 'profile'
            },
        },
        {
            path: '/login',
            name: 'login',
            component: Login,
            beforeEnter: () => {
                // Refuse navigation to /login if user already logged in
                // if (useUserStore().user?.username) {
                //     return false
                // }
            },
        },
        {
            path: '/admin/login',
            name: 'admin-login',
            component: AdminLogin,
        },
        {
            path: '/profile',
            name: 'profile',
            component: Profile
        },
        {
            path: '/about',
            name: 'about',
            // route level code-splitting
            // this generates a separate chunk (About.[hash].js) for this route
            // which is lazy-loaded when the route is visited.
            component: () => import('../views/AboutView.vue')
        },
        {
            path: '/organizations',
            name: 'organizations',
            component: OrganizationList
        },
        {
            path: '/organizations/:organization',
            name: 'organization',
            component: OrganizationItem
        },
        {
            path: '/organizations/:organization/workspaces',
            name: 'workspaces',
            component: WorkspaceList,
        },
    ]
})

router.beforeEach(async (to, _) => {
    switch (to.fullPath) {
        case "/oauth/github/login":
        case "/logout":
            window.location.replace(to.fullPath)
            return false
    }
    return true
})

export default router
