<script setup>
import { RouterView } from 'vue-router'
import { isAuthenticated } from './services/axios.js'
import { onMounted, ref } from 'vue'

const isLoggedIn = ref(false)

onMounted(() => {
	isLoggedIn.value = isAuthenticated()
})
</script>

<template>
	<div id="app">
		<!-- Layout per pagine pubbliche (login) -->
		<div v-if="!isLoggedIn" class="public-layout">
			<RouterView />
		</div>
		
		<!-- Layout chat per utenti autenticati -->
		<div v-else class="chat-layout">
			<RouterView />
		</div>
	</div>
</template>

<style>
/* Reset e stili base */
* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

body {
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', sans-serif;
	background-color: #f8f9fa;
	overflow-x: hidden;
}

#app {
	min-height: 100vh;
	display: flex;
	flex-direction: column;
}

/* Layout pubblico (login) */
.public-layout {
	min-height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
}

/* Layout chat */
.chat-layout {
	min-height: 100vh;
	display: flex;
	flex-direction: column;
	background-color: #f0f2f5;
}

/* Utility classes */
.flex {
	display: flex;
}

.flex-column {
	flex-direction: column;
}

.items-center {
	align-items: center;
}

.justify-center {
	justify-content: center;
}

.justify-between {
	justify-content: space-between;
}

.w-full {
	width: 100%;
}

.h-full {
	height: 100%;
}

.min-h-screen {
	min-height: 100vh;
}

/* Stili per i componenti chat */
.chat-container {
	display: flex;
	height: 100vh;
	background-color: #f0f2f5;
}

.sidebar {
	width: 350px;
	background-color: white;
	border-right: 1px solid #e1e5e9;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

.chat-main {
	flex: 1;
	display: flex;
	flex-direction: column;
	background-color: white;
	overflow: hidden;
}

/* Scrollbar personalizzata */
::-webkit-scrollbar {
	width: 6px;
}

::-webkit-scrollbar-track {
	background: #f1f1f1;
}

::-webkit-scrollbar-thumb {
	background: #c1c1c1;
	border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
	background: #a8a8a8;
}

/* Responsive */
@media (max-width: 768px) {
	body {
		font-size: 14px;
	}
	
	.sidebar {
		width: 100%;
		position: fixed;
		top: 0;
		left: 0;
		height: 100vh;
		z-index: 1000;
		transform: translateX(-100%);
		transition: transform 0.3s ease;
	}
	
	.sidebar.open {
		transform: translateX(0);
	}
	
	.chat-main {
		width: 100%;
	}
}

/* Animazioni */
.fade-enter-active, .fade-leave-active {
	transition: opacity 0.3s ease;
}

.fade-enter-from, .fade-leave-to {
	opacity: 0;
}

.slide-enter-active, .slide-leave-active {
	transition: transform 0.3s ease;
}

.slide-enter-from {
	transform: translateX(-100%);
}

.slide-leave-to {
	transform: translateX(100%);
}
</style>
