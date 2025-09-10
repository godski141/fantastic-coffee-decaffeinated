<script>
export default {
	name: 'Message',
	props: {
		message: {
			type: Object,
			required: true
		},
		isOwn: {
			type: Boolean,
			default: false
		},
		showAvatar: {
			type: Boolean,
			default: true
		},
		showTime: {
			type: Boolean,
			default: true
		}
	},
	computed: {
		formattedTime() {
			if (!this.message.timestamp) return ''
			const date = new Date(this.message.timestamp)
			const now = new Date()
			const diff = now - date
			
			if (diff < 60000) return 'Ora'
			if (diff < 3600000) return `${Math.floor(diff / 60000)}m fa`
			if (diff < 86400000) return `${Math.floor(diff / 3600000)}h fa`
			if (diff < 604800000) return `${Math.floor(diff / 86400000)}g fa`
			return date.toLocaleDateString('it-IT', { day: '2-digit', month: '2-digit' })
		},
		
		senderInitial() {
			return this.message.senderName?.charAt(0).toUpperCase() || 'U'
		},
		
		messageType() {
			// Determina il tipo di messaggio per styling diverso
			if (this.message.type === 'reply') return 'reply'
			if (this.message.type === 'forward') return 'forward'
			if (this.message.type === 'image') return 'image'
			return 'text'
		}
	},
	methods: {
		handleReaction(emoji) {
			this.$emit('reaction', { messageId: this.message.id, emoji })
		},
		
		handleReply() {
			this.$emit('reply', this.message)
		},
		
		handleForward() {
			this.$emit('forward', this.message)
		},
		
		handleDelete() {
			this.$emit('delete', this.message)
		},
		
		handleMenuToggle() {
			this.$emit('menu-toggle', this.message)
		}
	}
}
</script>

<template>
	<div class="message-wrapper" :class="{ 'own-message': isOwn }">
		<!-- Avatar (solo per messaggi altrui) -->
		<div v-if="!isOwn && showAvatar" class="message-avatar">
			{{ senderInitial }}
		</div>
		
		<!-- Message Content -->
		<div class="message-content">
			<!-- Sender Name (solo per messaggi altrui nei gruppi) -->
			<div v-if="!isOwn && message.senderName" class="sender-name">
				{{ message.senderName }}
			</div>
			
			<!-- Message Bubble -->
			<div class="message-bubble" :class="[`message-${messageType}`, { 'own': isOwn }]">
				<!-- Reply Context -->
				<div v-if="message.replyTo" class="reply-context">
					<div class="reply-line"></div>
					<div class="reply-content">
						<div class="reply-sender">{{ message.replyTo.senderName }}</div>
						<div class="reply-text">{{ message.replyTo.content }}</div>
					</div>
				</div>
				
				<!-- Message Text/Content -->
				<div class="message-text">
					{{ message.content }}
				</div>
				
				<!-- Message Time -->
				<div v-if="showTime" class="message-time">
					{{ formattedTime }}
				</div>
			</div>
			
			<!-- Message Actions -->
			<div class="message-actions" v-if="isOwn">
				<button 
					@click="handleReaction('👍')" 
					class="action-btn"
					title="Reazione"
				>
					👍
				</button>
				<button 
					@click="handleReply" 
					class="action-btn"
					title="Rispondi"
				>
					↩️
				</button>
				<button 
					@click="handleForward" 
					class="action-btn"
					title="Inoltra"
				>
					↗️
				</button>
				<button 
					@click="handleDelete" 
					class="action-btn delete"
					title="Elimina"
				>
					🗑️
				</button>
			</div>
		</div>
	</div>
</template>

<style scoped>
.message-wrapper {
	display: flex;
	align-items: flex-end;
	gap: 0.5rem;
	margin-bottom: 0.5rem;
	width: 100%;
}

.message-wrapper.own-message {
	flex-direction: row-reverse;
}

.message-avatar {
	width: 32px;
	height: 32px;
	border-radius: 50%;
	background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
	color: white;
	display: flex;
	align-items: center;
	justify-content: center;
	font-weight: 600;
	font-size: 0.8rem;
	flex-shrink: 0;
}

.message-content {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
	max-width: 70%;
}

.own-message .message-content {
	align-items: flex-end;
}

.sender-name {
	font-size: 0.8rem;
	font-weight: 600;
	color: #667eea;
	margin-bottom: 0.25rem;
}

.message-bubble {
	background: white;
	border-radius: 18px;
	padding: 0.75rem 1rem;
	box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
	position: relative;
	word-wrap: break-word;
}

.message-bubble.own {
	background: #667eea;
	color: white;
}

.message-bubble.message-reply {
	border-left: 3px solid #667eea;
}

.message-bubble.message-forward {
	border-left: 3px solid #ffc107;
}

.reply-context {
	display: flex;
	gap: 0.5rem;
	margin-bottom: 0.5rem;
	padding-bottom: 0.5rem;
	border-bottom: 1px solid rgba(0, 0, 0, 0.1);
}

.own .reply-context {
	border-bottom-color: rgba(255, 255, 255, 0.3);
}

.reply-line {
	width: 3px;
	background: #667eea;
	border-radius: 2px;
	flex-shrink: 0;
}

.reply-content {
	flex: 1;
}

.reply-sender {
	font-size: 0.75rem;
	font-weight: 600;
	color: #667eea;
	margin-bottom: 0.25rem;
}

.own .reply-sender {
	color: rgba(255, 255, 255, 0.8);
}

.reply-text {
	font-size: 0.8rem;
	color: #6c757d;
	line-height: 1.3;
}

.own .reply-text {
	color: rgba(255, 255, 255, 0.7);
}

.message-text {
	font-size: 0.95rem;
	line-height: 1.4;
	white-space: pre-wrap;
}

.message-time {
	font-size: 0.75rem;
	opacity: 0.7;
	margin-top: 0.25rem;
	align-self: flex-end;
}

.message-actions {
	display: flex;
	gap: 0.25rem;
	margin-top: 0.25rem;
	opacity: 0;
	transition: opacity 0.2s ease;
}

.message-wrapper:hover .message-actions {
	opacity: 1;
}

.action-btn {
	background: none;
	border: none;
	cursor: pointer;
	padding: 0.25rem;
	border-radius: 4px;
	font-size: 0.8rem;
	transition: all 0.2s ease;
	opacity: 0.7;
}

.action-btn:hover {
	background: rgba(0, 0, 0, 0.1);
	opacity: 1;
}

.action-btn.delete:hover {
	background: rgba(220, 53, 69, 0.1);
	color: #dc3545;
}

/* Responsive */
@media (max-width: 768px) {
	.message-content {
		max-width: 85%;
	}
	
	.message-avatar {
		width: 28px;
		height: 28px;
		font-size: 0.7rem;
	}
	
	.message-bubble {
		padding: 0.6rem 0.8rem;
	}
	
	.message-text {
		font-size: 0.9rem;
	}
	
	.message-actions {
		opacity: 1; /* Sempre visibili su mobile */
	}
}

/* Animazioni */
.message-wrapper {
	animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
	from {
		opacity: 0;
		transform: translateY(10px);
	}
	to {
		opacity: 1;
		transform: translateY(0);
	}
}

/* Stati speciali */
.message-bubble.sending {
	opacity: 0.7;
}

.message-bubble.failed {
	border: 1px solid #dc3545;
	background: #f8d7da;
}

.message-bubble.failed.own {
	background: rgba(220, 53, 69, 0.1);
	border-color: #dc3545;
}
</style>
