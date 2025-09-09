import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import { isAuthenticated } from '../services/axios.js'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		// Rotte pubbliche
		{
			path: '/',
			name: 'login',
			component: LoginView,
			meta: { requiresAuth: false }
		},
		{
			path: '/login',
			name: 'login-page',
			component: LoginView,
			meta: { requiresAuth: false }
		},
		
		// Rotte protette (richiedono autenticazione)
		{
			path: '/home',
			name: 'home',
			component: HomeView,
			meta: { requiresAuth: true }
		},
		
		// Redirect per rotte non trovate
		{
			path: '/:pathMatch(.*)*',
			redirect: '/'
		}
	]
})

// Guard per verificare l'autenticazione
router.beforeEach((to, from, next) => {
	const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
	const authenticated = isAuthenticated()
	
	if (requiresAuth && !authenticated) {
		// Se la rotta richiede autenticazione ma l'utente non è autenticato
		next('/login')
	} else if (!requiresAuth && authenticated && (to.path === '/' || to.path === '/login')) {
		// Se l'utente è già autenticato e sta cercando di andare al login
		next('/home')
	} else {
		// Altrimenti procedi normalmente
		next()
	}
})

export default router
