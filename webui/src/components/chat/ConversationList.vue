<script>
import { getUserData, logout } from '../../services/axios.js'

export default {
	name: 'ConversationList',
	props: {
		conversations: {
			type: Array,
			default: () => []
		},
		selectedConversation: {
			type: Object,
			default: null
		}
	},
	emits: ['conversation-selected', 'new-conversation', 'search-users', 'new-group'],
	data() {
		return {
			searchQuery: '',
			user: null
		}
	},
	computed: {
		filteredConversations() {
			if (!this.conversations || !Array.isArray(this.conversations)) return []
			if (!this.searchQuery) return this.conversations
			return this.conversations.filter(conv => 
				conv.Name?.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
				conv.LastMessage?.toLowerCase().includes(this.searchQuery.toLowerCase())
			)
		},
		selectedConversationId() {
			return this.selectedConversation?.convID || null
		}
	},
	mounted() {
		this.user = getUserData()
	},
	methods: {
		selectConversation(conversation) {
			this.$emit('conversation-selected', conversation)
		},
		
		handleNewConversation() {
			this.$emit('new-conversation')
		},
		
		handleSearchUsers() {
			this.$emit('search-users')
		},
		
		handleNewGroup() {
			this.$emit('new-group')
		},
		
		handleLogout() {
			logout()
			this.$router.push('/login')
		},
		
		formatLastMessage(message) {
			if (!message) return 'Nessun messaggio'
			return message.length > 50 ? message.substring(0, 50) + '...' : message
		},
		
		formatTime(timestamp) {
			if (!timestamp) return ''
			const date = new Date(timestamp)
			const now = new Date()
			const diff = now - date
			
			if (diff < 60000) return 'Ora'
			if (diff < 3600000) return `${Math.floor(diff / 60000)}m`
			if (diff < 86400000) return `${Math.floor(diff / 3600000)}h`
			if (diff < 604800000) return `${Math.floor(diff / 86400000)}g`
			return date.toLocaleDateString('it-IT', { day: '2-digit', month: '2-digit' })
		},
		
		getConversationAvatar(conversation) {
			if (conversation.isGroup) {
				return conversation.Name?.charAt(0).toUpperCase() || 'G'
			}
			return conversation.Name?.charAt(0).toUpperCase() || 'U'
		},
		
		getConversationName(conversation) {
			return conversation.Name || 'Conversazione senza nome'
		}
	}
}
</script>

<template>
	<div class="conversation-list">
		<!-- Header -->
		<div class="conversation-header">
			<div class="user-info">
				<div class="user-avatar">
					{{ user?.username?.charAt(0).toUpperCase() }}
				</div>
				<div class="user-details">
					<div class="username">{{ user?.username }}</div>
					<div class="user-status">Online</div>
				</div>
			</div>
			<button @click="handleLogout" class="logout-btn" title="Logout">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
					<polyline points="16,17 21,12 16,7"></polyline>
					<line x1="21" y1="12" x2="9" y2="12"></line>
				</svg>
			</button>
		</div>

		<!-- Search Bar -->
		<div class="search-container">
			<div class="search-input">
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="11" cy="11" r="8"></circle>
					<path d="M21 21l-4.35-4.35"></path>
				</svg>
				<input 
					v-model="searchQuery"
					type="text" 
					placeholder="Cerca conversazioni..." 
				/>
			</div>
		</div>

		<!-- Actions -->
		<div class="actions-container">
			<div class="actions-header">
				<h3>Conversazioni</h3>
				<div class="action-buttons">
					<button @click="handleSearchUsers" class="action-btn" title="Cerca utenti">
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="11" cy="11" r="8"></circle>
							<path d="M21 21l-4.35-4.35"></path>
						</svg>
					</button>
					<button @click="handleNewConversation" class="action-btn primary" title="Nuova conversazione">
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="5" x2="12" y2="19"></line>
							<line x1="5" y1="12" x2="19" y2="12"></line>
						</svg>
					</button>
					<button @click="handleNewGroup" class="action-btn" title="Nuovo gruppo">
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
							<circle cx="9" cy="7" r="4"></circle>
							<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
							<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
						</svg>
					</button>
				</div>
			</div>
		</div>

		<!-- Conversations List -->
		<div class="conversations-container">
			<!-- Empty State -->
			<div v-if="!filteredConversations || filteredConversations.length === 0" class="empty-state">
				<div class="empty-icon">💬</div>
				<p v-if="searchQuery">Nessuna conversazione trovata</p>
				<p v-else>Nessuna conversazione</p>
				<small v-if="!searchQuery">Inizia una nuova chat per vedere le conversazioni qui</small>
			</div>

			<!-- Conversations List -->
			<div v-if="filteredConversations && filteredConversations.length > 0" class="conversations-list">
				<div 
					v-for="conversation in filteredConversations" 
					:key="conversation.convID"
					@click="selectConversation(conversation)"
					class="conversation-item"
					:class="{ active: selectedConversationId === conversation.convID }"
				>
					<div class="conversation-avatar">
						{{ getConversationAvatar(conversation) }}
					</div>
					<div class="conversation-content">
						<div class="conversation-header">
							<div class="conversation-name">{{ getConversationName(conversation) }}</div>
							<div class="conversation-time">{{ formatTime(conversation.LastMessageTime) }}</div>
						</div>
						<div class="conversation-preview">
							<span class="last-message">{{ formatLastMessage(conversation.LastMessage) }}</span>
							<div v-if="conversation.unreadCount > 0" class="unread-badge">
								{{ conversation.unreadCount > 99 ? '99+' : conversation.unreadCount }}
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.conversation-list {
	height: 100%;
	display: flex;
	flex-direction: column;
	background-color: white;
}

/* Header */
.conversation-header {
	padding: 1rem;
	border-bottom: 1px solid #e1e5e9;
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.user-info {
	display: flex;
	align-items: center;
	gap: 0.75rem;
}

.user-avatar {
	width: 40px;
	height: 40px;
	border-radius: 50%;
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	color: white;
	display: flex;
	align-items: center;
	justify-content: center;
	font-weight: 600;
	font-size: 1.1rem;
}

.user-details {
	flex: 1;
}

.username {
	font-weight: 600;
	color: #333;
	font-size: 0.95rem;
}

.user-status {
	font-size: 0.8rem;
	color: #28a745;
}

.logout-btn {
	background: none;
	border: none;
	color: #6c757d;
	cursor: pointer;
	padding: 0.5rem;
	border-radius: 6px;
	transition: all 0.2s ease;
}

.logout-btn:hover {
	background-color: #f8f9fa;
	color: #dc3545;
}

/* Search */
.search-container {
	padding: 1rem;
	border-bottom: 1px solid #e1e5e9;
}

.search-input {
	position: relative;
	display: flex;
	align-items: center;
}

.search-input svg {
	position: absolute;
	left: 12px;
	color: #6c757d;
}

.search-input input {
	width: 100%;
	padding: 0.75rem 0.75rem 0.75rem 2.5rem;
	border: 1px solid #e1e5e9;
	border-radius: 20px;
	font-size: 0.9rem;
	outline: none;
	transition: all 0.2s ease;
}

.search-input input:focus {
	border-color: #667eea;
	box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

/* Actions */
.actions-container {
	padding: 1rem;
	border-bottom: 1px solid #e1e5e9;
}

.actions-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.actions-header h3 {
	margin: 0;
	font-size: 1rem;
	color: #333;
	font-weight: 600;
}

.action-buttons {
	display: flex;
	gap: 0.5rem;
}

.action-btn {
	background: #f8f9fa;
	color: #6c757d;
	border: none;
	border-radius: 50%;
	width: 32px;
	height: 32px;
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	transition: all 0.2s ease;
}

.action-btn:hover {
	background: #e9ecef;
	color: #495057;
}

.action-btn.primary {
	background: #667eea;
	color: white;
}

.action-btn.primary:hover {
	background: #5a6fd8;
	transform: scale(1.05);
}

/* Conversations Container */
.conversations-container {
	flex: 1;
	overflow-y: auto;
}

/* Loading State */
.loading-state {
	padding: 2rem;
	text-align: center;
	color: #6c757d;
}

.spinner {
	width: 32px;
	height: 32px;
	border: 3px solid #f3f3f3;
	border-top: 3px solid #667eea;
	border-radius: 50%;
	animation: spin 1s linear infinite;
	margin: 0 auto 1rem;
}

@keyframes spin {
	0% { transform: rotate(0deg); }
	100% { transform: rotate(360deg); }
}

/* Error State */
.error-state {
	padding: 2rem;
	text-align: center;
	color: #dc3545;
}

.error-icon {
	font-size: 2rem;
	margin-bottom: 1rem;
}

.retry-btn {
	background: #667eea;
	color: white;
	border: none;
	padding: 0.5rem 1rem;
	border-radius: 6px;
	cursor: pointer;
	margin-top: 1rem;
	transition: all 0.2s ease;
}

.retry-btn:hover {
	background: #5a6fd8;
}

/* Empty State */
.empty-state {
	padding: 2rem;
	text-align: center;
	color: #6c757d;
}

.empty-icon {
	font-size: 3rem;
	margin-bottom: 1rem;
}

.empty-state p {
	margin: 0.5rem 0;
	font-weight: 500;
}

.empty-state small {
	font-size: 0.8rem;
}

/* Conversations List */
.conversations-list {
	padding: 0;
}

.conversation-item {
	display: flex;
	align-items: center;
	padding: 1rem;
	cursor: pointer;
	transition: all 0.2s ease;
	border-bottom: 1px solid #f8f9fa;
}

.conversation-item:hover {
	background-color: #f8f9fa;
}

.conversation-item.active {
	background-color: #e3f2fd;
	border-left: 3px solid #667eea;
}

.conversation-avatar {
	width: 48px;
	height: 48px;
	border-radius: 50%;
	background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
	color: white;
	display: flex;
	align-items: center;
	justify-content: center;
	font-weight: 600;
	font-size: 1.2rem;
	margin-right: 0.75rem;
	flex-shrink: 0;
}

.conversation-content {
	flex: 1;
	min-width: 0;
}

.conversation-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 0.25rem;
}

.conversation-name {
	font-weight: 600;
	color: #333;
	font-size: 0.95rem;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.conversation-time {
	font-size: 0.8rem;
	color: #6c757d;
	flex-shrink: 0;
}

.conversation-preview {
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.last-message {
	color: #6c757d;
	font-size: 0.85rem;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
	flex: 1;
}

.unread-badge {
	background: #667eea;
	color: white;
	border-radius: 10px;
	padding: 0.25rem 0.5rem;
	font-size: 0.75rem;
	font-weight: 600;
	margin-left: 0.5rem;
	flex-shrink: 0;
}

/* Responsive */
@media (max-width: 768px) {
	.conversation-item {
		padding: 0.75rem;
	}
	
	.conversation-avatar {
		width: 40px;
		height: 40px;
		font-size: 1rem;
	}
	
	.conversation-name {
		font-size: 0.9rem;
	}
	
	.last-message {
		font-size: 0.8rem;
	}
}
</style>
