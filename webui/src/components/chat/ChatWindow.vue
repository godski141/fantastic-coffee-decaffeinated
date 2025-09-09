<script>
import { getMessages, sendMessage } from '../../services/axios.js'

export default {
	name: 'ChatWindow',
	props: {
		conversation: {
			type: Object,
			default: null
		}
	},
	data() {
		return {
			messages: [],
			loading: false,
			error: null,
			newMessage: '',
			sending: false,
			scrollToBottom: true
		}
	},
	watch: {
		conversation: {
			handler(newConv) {
				if (newConv) {
					this.loadMessages()
				} else {
					this.messages = []
				}
			},
			immediate: true
		}
	},
	mounted() {
		// Auto-scroll to bottom when component mounts
		this.$nextTick(() => {
			this.scrollToBottomOfMessages()
		})
	},
	updated() {
		// Auto-scroll to bottom when messages change
		if (this.scrollToBottom) {
			this.$nextTick(() => {
				this.scrollToBottomOfMessages()
			})
		}
	},
	methods: {
		async loadMessages() {
			if (!this.conversation?.ConvID) return
			
			this.loading = true
			this.error = null
			try {
				this.messages = await getMessages(this.conversation.ConvID)
			} catch (error) {
				console.error('Errore nel caricamento messaggi:', error)
				this.error = 'Errore nel caricamento dei messaggi'
			} finally {
				this.loading = false
			}
		},
		
		async sendNewMessage() {
			if (!this.newMessage.trim() || this.sending || !this.conversation?.ConvID) return
			
			const messageText = this.newMessage.trim()
			this.newMessage = ''
			this.sending = true
			
			try {
				const newMessage = await sendMessage(this.conversation.ConvID, messageText)
				this.messages.push(newMessage)
				this.scrollToBottom = true
			} catch (error) {
				console.error('Errore nell\'invio messaggio:', error)
				this.error = 'Errore nell\'invio del messaggio'
				// Ripristina il messaggio in caso di errore
				this.newMessage = messageText
			} finally {
				this.sending = false
			}
		},
		
		handleKeyPress(event) {
			if (event.key === 'Enter' && !event.shiftKey) {
				event.preventDefault()
				this.sendNewMessage()
			}
		},
		
		scrollToBottomOfMessages() {
			const messagesContainer = this.$refs.messagesContainer
			if (messagesContainer) {
				messagesContainer.scrollTop = messagesContainer.scrollHeight
			}
		},
		
		onScroll() {
			const messagesContainer = this.$refs.messagesContainer
			if (messagesContainer) {
				const { scrollTop, scrollHeight, clientHeight } = messagesContainer
				// Se l'utente è vicino al fondo, mantieni l'auto-scroll
				this.scrollToBottom = scrollTop + clientHeight >= scrollHeight - 100
			}
		},
		
		formatMessageTime(timestamp) {
			if (!timestamp) return ''
			const date = new Date(timestamp)
			const now = new Date()
			const diff = now - date
			
			if (diff < 60000) return 'Ora'
			if (diff < 3600000) return `${Math.floor(diff / 60000)}m fa`
			if (diff < 86400000) return `${Math.floor(diff / 3600000)}h fa`
			if (diff < 604800000) return `${Math.floor(diff / 86400000)}g fa`
			return date.toLocaleDateString('it-IT', { day: '2-digit', month: '2-digit' })
		},
		
		isMyMessage(message) {
			// Recupera l'ID utente dal localStorage
			const currentUserId = localStorage.getItem('userId')
			return message.SenderID === currentUserId
		}
	}
}
</script>

<template>
	<div class="chat-window">
		<!-- Chat Header -->
		<div class="chat-header">
			<div v-if="conversation" class="chat-info">
				<div class="chat-avatar">
					{{ conversation.Name?.charAt(0).toUpperCase() || 'C' }}
				</div>
				<div class="chat-details">
					<div class="chat-name">{{ conversation.Name || 'Conversazione' }}</div>
					<div class="chat-status">Online</div>
				</div>
			</div>
			
			<div v-else class="welcome-info">
				<h2>Benvenuto in WasaTEXT!</h2>
				<p>Seleziona una conversazione per iniziare a chattare</p>
			</div>
		</div>

		<!-- Messages Area -->
		<div class="messages-container">
			<!-- Loading State -->
			<div v-if="loading" class="loading-state">
				<div class="spinner"></div>
				<p>Caricamento messaggi...</p>
			</div>

			<!-- Error State -->
			<div v-else-if="error" class="error-state">
				<div class="error-icon">⚠️</div>
				<p>{{ error }}</p>
				<button @click="loadMessages" class="retry-btn">Riprova</button>
			</div>

			<!-- Welcome State -->
			<div v-else-if="!conversation" class="welcome-state">
				<div class="welcome-icon">💬</div>
				<h3>Inizia una conversazione</h3>
				<p>Seleziona una chat dalla sidebar o creane una nuova</p>
				<div class="welcome-features">
					<div class="feature">
						<span class="feature-icon">📱</span>
						<span>Messaggi istantanei</span>
					</div>
					<div class="feature">
						<span class="feature-icon">👥</span>
						<span>Gruppi e chat private</span>
					</div>
					<div class="feature">
						<span class="feature-icon">😊</span>
						<span>Reazioni e emoji</span>
					</div>
				</div>
			</div>

			<!-- Messages List -->
			<div 
				v-else
				ref="messagesContainer"
				class="messages-list"
				@scroll="onScroll"
			>
				<div v-if="messages.length === 0" class="empty-messages">
					<div class="empty-icon">💬</div>
					<p>Nessun messaggio ancora</p>
					<small>Invia il primo messaggio per iniziare la conversazione</small>
				</div>
				
				<div v-else class="messages">
					<div 
						v-for="message in messages" 
						:key="message.MessageID"
						class="message-wrapper"
						:class="{ 'my-message': isMyMessage(message) }"
					>
						<div class="message">
							<div class="message-content">
								<div class="message-text">{{ message.Content }}</div>
								<div class="message-time">{{ formatMessageTime(message.Timestamp) }}</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Message Input -->
		<div v-if="conversation" class="message-input-container">
			<div class="message-input-wrapper">
				<div class="message-input">
					<textarea
						v-model="newMessage"
						@keypress="handleKeyPress"
						placeholder="Scrivi un messaggio..."
						:disabled="sending"
						rows="1"
						ref="messageInput"
					></textarea>
					<button 
						@click="sendNewMessage"
						:disabled="!newMessage.trim() || sending"
						class="send-button"
					>
						<svg v-if="!sending" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="22" y1="2" x2="11" y2="13"></line>
							<polygon points="22,2 15,22 11,13 2,9 22,2"></polygon>
						</svg>
						<div v-else class="sending-spinner"></div>
					</button>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.chat-window {
	height: 100%;
	display: flex;
	flex-direction: column;
	background-color: white;
}

/* Chat Header */
.chat-header {
	padding: 1rem;
	border-bottom: 1px solid #e1e5e9;
	display: flex;
	align-items: center;
	gap: 1rem;
	background-color: white;
	flex-shrink: 0;
}

.chat-info {
	display: flex;
	align-items: center;
	gap: 0.75rem;
}

.chat-avatar {
	width: 40px;
	height: 40px;
	border-radius: 50%;
	background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
	color: white;
	display: flex;
	align-items: center;
	justify-content: center;
	font-weight: 600;
}

.chat-details {
	flex: 1;
}

.chat-name {
	font-weight: 600;
	color: #333;
}

.chat-status {
	font-size: 0.8rem;
	color: #28a745;
}

.welcome-info h2 {
	margin: 0;
	color: #333;
	font-size: 1.5rem;
}

.welcome-info p {
	margin: 0.25rem 0 0 0;
	color: #6c757d;
	font-size: 0.9rem;
}

/* Messages Container */
.messages-container {
	flex: 1;
	overflow: hidden;
	display: flex;
	flex-direction: column;
}

/* Loading State */
.loading-state {
	padding: 2rem;
	text-align: center;
	color: #6c757d;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	height: 100%;
}

.spinner {
	width: 32px;
	height: 32px;
	border: 3px solid #f3f3f3;
	border-top: 3px solid #667eea;
	border-radius: 50%;
	animation: spin 1s linear infinite;
	margin-bottom: 1rem;
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
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	height: 100%;
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

/* Welcome State */
.welcome-state {
	padding: 2rem;
	text-align: center;
	color: #6c757d;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	height: 100%;
}

.welcome-icon {
	font-size: 4rem;
	margin-bottom: 1.5rem;
}

.welcome-state h3 {
	color: #333;
	margin-bottom: 1rem;
	font-size: 1.5rem;
}

.welcome-state p {
	margin-bottom: 2rem;
}

.welcome-features {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
	gap: 1rem;
	max-width: 500px;
}

.feature {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.5rem;
	padding: 1rem;
	background: #f8f9fa;
	border-radius: 8px;
}

.feature-icon {
	font-size: 1.5rem;
}

/* Messages List */
.messages-list {
	flex: 1;
	overflow-y: auto;
	padding: 1rem;
	background: #f8f9fa;
}

.empty-messages {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	height: 100%;
	color: #6c757d;
}

.empty-icon {
	font-size: 3rem;
	margin-bottom: 1rem;
}

.empty-messages p {
	margin: 0.5rem 0;
	font-weight: 500;
}

.empty-messages small {
	font-size: 0.8rem;
}

.messages {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.message-wrapper {
	display: flex;
	width: 100%;
}

.message-wrapper.my-message {
	justify-content: flex-end;
}

.message {
	max-width: 70%;
	background: white;
	border-radius: 18px;
	padding: 0.75rem 1rem;
	box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.message-wrapper.my-message .message {
	background: #667eea;
	color: white;
}

.message-content {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
}

.message-text {
	font-size: 0.95rem;
	line-height: 1.4;
	word-wrap: break-word;
}

.message-time {
	font-size: 0.75rem;
	opacity: 0.7;
	align-self: flex-end;
}

/* Message Input */
.message-input-container {
	padding: 1rem;
	border-top: 1px solid #e1e5e9;
	background-color: white;
	flex-shrink: 0;
}

.message-input-wrapper {
	max-width: 100%;
}

.message-input {
	display: flex;
	align-items: flex-end;
	gap: 0.75rem;
	background: #f8f9fa;
	border-radius: 24px;
	padding: 0.75rem 1rem;
	border: 1px solid #e1e5e9;
	transition: all 0.2s ease;
}

.message-input:focus-within {
	border-color: #667eea;
	box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.message-input textarea {
	flex: 1;
	border: none;
	background: none;
	outline: none;
	resize: none;
	font-size: 0.95rem;
	line-height: 1.4;
	max-height: 120px;
	min-height: 20px;
	font-family: inherit;
}

.message-input textarea::placeholder {
	color: #6c757d;
}

.send-button {
	background: #667eea;
	color: white;
	border: none;
	border-radius: 50%;
	width: 36px;
	height: 36px;
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	transition: all 0.2s ease;
	flex-shrink: 0;
}

.send-button:hover:not(:disabled) {
	background: #5a6fd8;
	transform: scale(1.05);
}

.send-button:disabled {
	background: #6c757d;
	cursor: not-allowed;
	transform: none;
}

.sending-spinner {
	width: 16px;
	height: 16px;
	border: 2px solid rgba(255, 255, 255, 0.3);
	border-top: 2px solid white;
	border-radius: 50%;
	animation: spin 1s linear infinite;
}

/* Responsive */
@media (max-width: 768px) {
	.chat-header {
		padding: 0.75rem;
	}
	
	.messages-list {
		padding: 0.75rem;
	}
	
	.message-input-container {
		padding: 0.75rem;
	}
	
	.message {
		max-width: 85%;
	}
	
	.welcome-features {
		grid-template-columns: 1fr;
	}
}
</style>
