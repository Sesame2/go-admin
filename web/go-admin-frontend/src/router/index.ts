import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { ElMessage } from 'element-plus'

const routes: RouteRecordRaw[] = [
    {
        path: '/login',
        name: 'Login',
        component: () => import('../views/Login.vue'),
        meta: { requiresAuth: false }
    },
    {
        path: '/',
        component: () => import('../components/Layout.vue'),
        meta: { requiresAuth: true },
        children: [
            {
                path: '',
                name: 'KnowledgeBase',
                component: () => import('../views/KnowledgeBase.vue')
            },
            {
                path: 'knowledge-base/:kbId/documents',
                name: 'Document',
                component: () => import('../views/Document.vue')
            },
            {
                path: 'knowledge-base/:kbId/retrieval',
                name: 'Retrieval',
                component: () => import('../views/Retrieval.vue')
            }
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// Navigation guard
router.beforeEach((to, _from, next) => {
    const token = localStorage.getItem('token')

    if (to.meta.requiresAuth && !token) {
        ElMessage.warning('请先登录')
        next('/login')
    } else if (to.path === '/login' && token) {
        next('/')
    } else {
        next()
    }
})

export default router
