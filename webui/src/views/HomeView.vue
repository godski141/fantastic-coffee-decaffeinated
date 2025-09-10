<script>
import { 
	getUserData, 
	logout, 
	getMyConversations, 
	startConversation, 
	createGroup,
	searchUsers 
} from '../services/axios.js'
import ConversationList from '../components/chat/ConversationList.vue'
import ChatWindow from '../components/chat/ChatWindow.vue'

export default {
	name: 'HomeView',
	components: {
		ConversationList,
		ChatWindow
	},
	data() {
		return {
			// User state
			user: null,
			
			// UI state
			sidebarOpen: false,
			loading: false,
			error: null,
			
			// Conversations state
			conversations: [],
			selectedConversation: null,
			
			// Modals state
			showSearchDialog: false,
			showNewGroupDialog: false,
			showAddMemberDialog: false,
			
			// Search state
			searchQuery: '',
			searchResults: [],
			searching: false,
			
			// Group creation state
			groupName: '',
			selectedMembers: [],
			
			// Auto-refresh
			refreshInterval: null
		}
	},
	computed: {
		filteredConversations() {
			if (!this.searchQuery) return this.conversations
			return this.conversations.filter(conv => 
				conv.Name?.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
				conv.LastMessage?.toLowerCase().includes(this.searchQuery.toLowerCase())
			)
		}
	},
	async mounted() {
		this.user = getUserData()
		if (!this.user) {
			this.$router.push('/login')
			return
		}
		
		await this.initializeApp()
		this.startAutoRefresh()
	},
	beforeUnmount() {
		if (this.refreshInterval) {
			clearInterval(this.refreshInterval)
		}
	},
	methods: {
		// ===========================================
		// INITIALIZATION
		// ===========================================
		async initializeApp() {
			this.loading = true
			this.error = null
			
			try {
				await this.loadConversations()
			} catch (error) {
				console.error('Errore nell\'inizializzazione:', error)
				this.error = 'Errore nel caricamento dell\'app'
			} finally {
				this.loading = false
			}
		},
		
		// ===========================================
		// CONVERSATIONS MANAGEMENT
		// ===========================================
		async loadConversations() {
			try {
				this.conversations = await getMyConversations()
			} catch (error) {
				console.error('Errore nel caricamento conversazioni:', error)
				throw error
			}
		},
		
		selectConversation(conversation) {
			this.selectedConversation = conversation
			// Chiudi sidebar su mobile quando selezioni una conversazione
			if (window.innerWidth <= 768) {
				this.sidebarOpen = false
			}
		},
		
		// ===========================================
		// UI MANAGEMENT
		// ===========================================
		toggleSidebar() {
			this.sidebarOpen = !this.sidebarOpen
		},
		
		// ===========================================
		// MODAL MANAGEMENT
		// ===========================================
		handleNewConversation() {
			this.showSearchDialog = true
		},
		
		handleSearchUsers() {
			this.showSearchDialog = true
		},
		
		handleNewGroup() {
			this.showNewGroupDialog = true
		},
		
		closeDialogs() {
			this.showSearchDialog = false
			this.showNewGroupDialog = false
			this.showAddMemberDialog = false
			this.resetModalStates()
		},
		
		resetModalStates() {
			this.searchQuery = ''
			this.searchResults = []
			this.groupName = ''
			this.selectedMembers = []
		},
		
		// ===========================================
		// SEARCH FUNCTIONALITY
		// ===========================================
		async searchUsers() {
			if (!this.searchQuery.trim()) {
				this.searchResults = []
				return
			}
			
			this.searching = true
			try {
				this.searchResults = await searchUsers(this.searchQuery)
			} catch (error) {
				console.error('Errore nella ricerca utenti:', error)
				this.error = 'Errore nella ricerca utenti'
			} finally {
				this.searching = false
			}
		},
		
		// ===========================================
		// CONVERSATION CREATION
		// ===========================================
		async createDirectConversation(username) {
			try {
				const newConversation = await startConversation(username)
				await this.loadConversations()
				// Trova la conversazione appena creata nella lista aggiornata
				const createdConv = this.conversations.find(conv => conv.ConvID === newConversation.conversation_id)
				if (createdConv) {
					this.selectConversation(createdConv)
				} else {
					// Fallback: seleziona la prima conversazione se non troviamo quella creata
					this.selectConversation(this.conversations[0])
				}
				this.closeDialogs()
			} catch (error) {
				console.error('Errore nella creazione conversazione:', error)
				this.error = 'Errore nella creazione della conversazione'
			}
		},
		
		async createGroupConversation() {
			if (!this.groupName.trim() || this.selectedMembers.length === 0) {
				this.error = 'Nome gruppo e almeno un membro sono richiesti'
				return
			}
			
			try {
				const newGroup = await createGroup(this.groupName, this.selectedMembers)
				await this.loadConversations()
				this.selectConversation(newGroup)
				this.closeDialogs()
			} catch (error) {
				console.error('Errore nella creazione gruppo:', error)
				this.error = 'Errore nella creazione del gruppo'
			}
		},
		
		// ===========================================
		// AUTO-REFRESH
		// ===========================================
		startAutoRefresh() {
			// Aggiorna le conversazioni ogni 30 secondi
			this.refreshInterval = setInterval(async () => {
				try {
					await this.loadConversations()
				} catch (error) {
					console.error('Errore nel refresh automatico:', error)
				}
			}, 30000)
		},
		
		// ===========================================
		// ERROR HANDLING
		// ===========================================
		clearError() {
			this.error = null
		},
		
		// ===========================================
		// MEMBER MANAGEMENT
		// ===========================================
		addMember(username) {
			if (!this.selectedMembers.includes(username)) {
				this.selectedMembers.push(username)
			}
		},
		
		removeMember(username) {
			const index = this.selectedMembers.indexOf(username)
			if (index > -1) {
				this.selectedMembers.splice(index, 1)
			}
		},
		
		// ===========================================
		// LOGOUT
		// ===========================================
		handleLogout() {
			logout()
			this.$router.push('/login')
		}
	}
}
</script>

<template>
	<div class="chat-container">
		<!-- Global Error Message -->
		<div v-if="error" class="global-error">
			<div class="error-content">
				<span class="error-icon">⚠️</span>
				<span class="error-message">{{ error }}</span>
				<button @click="clearError" class="error-close">×</button>
			</div>
		</div>

		<!-- Loading Overlay -->
		<div v-if="loading" class="loading-overlay">
			<div class="loading-content">
				<div class="spinner"></div>
				<p>Caricamento...</p>
			</div>
		</div>

		<!-- Sidebar -->
		<div class="sidebar" :class="{ open: sidebarOpen }">
			<ConversationList 
				:conversations="conversations"
				:selected-conversation="selectedConversation"
				@conversation-selected="selectConversation"
				@new-conversation="handleNewConversation"
				@search-users="handleSearchUsers"
				@new-group="handleNewGroup"
			/>
		</div>

		<!-- Main Chat Area -->
		<div class="chat-main">
			<!-- Mobile Menu Button -->
			<button @click="toggleSidebar" class="menu-btn mobile-only">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="3" y1="6" x2="21" y2="6"></line>
					<line x1="3" y1="12" x2="21" y2="12"></line>
					<line x1="3" y1="18" x2="21" y2="18"></line>
				</svg>
			</button>
			
			
			<!-- Chat Window Component -->
			<ChatWindow :conversation="selectedConversation" />
		</div>



		<!-- Search Dialog Modal -->
		<div v-if="showSearchDialog" style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 10000; display: flex; align-items: center; justify-content: center; padding: 20px;" @click="closeDialogs">
			<div style="background: white; border-radius: 12px; box-shadow: 0 20px 40px rgba(0,0,0,0.2); width: 100%; max-width: 500px; max-height: 80vh; overflow: hidden; position: relative;" @click.stop>
				<!-- Header -->
				<div style="padding: 20px; border-bottom: 1px solid #e1e5e9; display: flex; justify-content: space-between; align-items: center;">
					<h3 style="margin: 0; color: #333; font-size: 1.25rem;">Nuova Conversazione</h3>
					<button @click="closeDialogs" style="background: none; border: none; font-size: 1.5rem; color: #6c757d; cursor: pointer; padding: 0; width: 24px; height: 24px; display: flex; align-items: center; justify-content: center;">×</button>
				</div>
				
				<!-- Content -->
				<div style="padding: 20px; max-height: 60vh; overflow-y: auto;">
					<!-- Search Input -->
					<div style="position: relative; margin-bottom: 20px;">
						<svg style="position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: #6c757d;" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="11" cy="11" r="8"></circle>
							<path d="M21 21l-4.35-4.35"></path>
						</svg>
						<input 
							v-model="searchQuery"
							@input="searchUsers"
							type="text" 
							placeholder="Cerca utenti..." 
							style="width: 100%; padding: 12px 12px 12px 40px; border: 1px solid #e1e5e9; border-radius: 8px; font-size: 0.95rem; outline: none; transition: all 0.2s ease;"
						/>
					</div>
					
					<!-- Search Loading -->
					<div v-if="searching" style="display: flex; align-items: center; gap: 0.5rem; color: #6c757d; padding: 20px; text-align: center; justify-content: center;">
						<div style="width: 16px; height: 16px; border: 2px solid #f3f3f3; border-top: 2px solid #667eea; border-radius: 50%; animation: spin 1s linear infinite;"></div>
						<span>Ricerca in corso...</span>
					</div>
					
					<!-- Search Results -->
					<div v-else-if="searchResults.length > 0" style="display: flex; flex-direction: column; gap: 8px;">
						<div 
							v-for="username in searchResults" 
							:key="username" 
							@click="createDirectConversation(username)" 
							style="display: flex; align-items: center; gap: 12px; padding: 12px; border-radius: 8px; cursor: pointer; transition: all 0.2s ease; border: 1px solid transparent;"
							@mouseover="$event.target.style.background = '#f8f9fa'; $event.target.style.borderColor = '#e1e5e9'"
							@mouseout="$event.target.style.background = 'white'; $event.target.style.borderColor = 'transparent'"
						>
							<div style="width: 40px; height: 40px; border-radius: 50%; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 1.1rem;">
								{{ username.charAt(0).toUpperCase() }}
							</div>
							<div style="flex: 1;">
								<div style="font-weight: 600; color: #333; margin-bottom: 4px;">{{ username }}</div>
								<div style="font-size: 0.85rem; color: #6c757d;">Crea conversazione</div>
							</div>
						</div>
					</div>
					
					<!-- No Results -->
					<div v-else-if="searchQuery && !searching" style="text-align: center; color: #6c757d; padding: 40px;">
						<p>Nessun utente trovato</p>
					</div>
					
					<!-- Empty State -->
					<div v-else style="text-align: center; color: #6c757d; padding: 40px;">
						<p>Inizia a digitare per cercare utenti</p>
					</div>
				</div>
			</div>
		</div>

		<!-- New Group Dialog Modal -->
		<div v-if="showNewGroupDialog" class="modal-overlay" @click="closeDialogs">
			<div class="modal" @click.stop>
				<div class="modal-header">
					<h3>Nuovo Gruppo</h3>
					<button @click="closeDialogs" class="modal-close">×</button>
				</div>
				<div class="modal-content">
					<div class="group-creation">
						<!-- Group Name -->
						<div class="form-group">
							<label for="groupName">Nome del gruppo</label>
							<input 
								v-model="groupName"
								id="groupName"
								type="text" 
								placeholder="Inserisci il nome del gruppo"
								maxlength="50"
							/>
						</div>
						
						<!-- Member Search -->
						<div class="form-group">
							<label>Aggiungi membri</label>
							<div class="search-input">
								<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="11" cy="11" r="8"></circle>
									<path d="M21 21l-4.35-4.35"></path>
								</svg>
								<input 
									v-model="searchQuery"
									@input="searchUsers"
									type="text" 
									placeholder="Cerca utenti da aggiungere..." 
								/>
							</div>
							
							<!-- Selected Members -->
							<div v-if="selectedMembers.length > 0" class="selected-members">
								<div 
									v-for="member in selectedMembers" 
									:key="member"
									class="selected-member"
								>
									<span>{{ member }}</span>
									<button @click="removeMember(member)" class="remove-member">×</button>
								</div>
							</div>
							
							<!-- Search Results -->
							<div v-if="searching" class="search-loading">
								<div class="spinner-small"></div>
								<span>Ricerca in corso...</span>
							</div>
							
							<div v-else-if="searchResults.length > 0" class="search-results">
								<div 
									v-for="username in searchResults" 
									:key="username"
									@click="addMember(username)"
									class="search-result-item"
									:class="{ disabled: selectedMembers.includes(username) }"
								>
									<div class="user-avatar">{{ username.charAt(0).toUpperCase() }}</div>
									<div class="user-info">
										<div class="username">{{ username }}</div>
										<div class="user-action">
											{{ selectedMembers.includes(username) ? 'Già aggiunto' : 'Aggiungi' }}
										</div>
									</div>
								</div>
							</div>
						</div>
						
						<!-- Create Button -->
						<div class="modal-actions">
							<button @click="closeDialogs" class="btn btn-secondary">Annulla</button>
							<button 
								@click="createGroupConversation"
								:disabled="!groupName.trim() || selectedMembers.length === 0"
								class="btn btn-primary"
							>
								Crea Gruppo
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
/* Chat Container */
.chat-container {
	display: flex;
	height: 100vh;
	background-color: #f0f2f5;
	width: 100%;
	overflow: hidden;
}

/* Sidebar */
.sidebar {
	width: 350px;
	min-width: 350px;
	max-width: 350px;
	background-color: white;
	border-right: 1px solid #e1e5e9;
	display: flex;
	flex-direction: column;
	overflow: hidden;
	flex-shrink: 0;
	position: relative;
	z-index: 1;
}

/* Chat Main */
.chat-main {
	flex: 1;
	display: flex;
	flex-direction: column;
	background-color: white;
	overflow: hidden;
	position: relative;
	min-width: 0;
}

.menu-btn {
	background: none;
	border: none;
	color: #6c757d;
	cursor: pointer;
	padding: 0.5rem;
	border-radius: 6px;
	display: none;
	position: absolute;
	top: 1rem;
	left: 1rem;
	z-index: 10;
}

.menu-btn:hover {
	background-color: #f8f9fa;
}

/* Responsive */
@media (max-width: 768px) {
	.menu-btn {
		display: block;
	}
	
	.sidebar {
		position: fixed;
		top: 0;
		left: 0;
		height: 100vh;
		z-index: 1000;
		transform: translateX(-100%);
		transition: transform 0.3s ease;
		width: 350px;
		min-width: 350px;
		max-width: 350px;
	}
	
	.sidebar.open {
		transform: translateX(0);
	}
	
	.chat-main {
		width: 100%;
	}
	
	.welcome-features {
		grid-template-columns: 1fr;
	}
}

/* Mobile only class */
.mobile-only {
	display: none;
}

@media (max-width: 768px) {
	.mobile-only {
		display: block;
	}
}

/* Global Error */
.global-error {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	background: #dc3545;
	color: white;
	z-index: 9999;
	padding: 0.75rem;
}

.error-content {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	max-width: 1200px;
	margin: 0 auto;
}

.error-icon {
	font-size: 1.2rem;
}

.error-message {
	flex: 1;
}

.error-close {
	background: none;
	border: none;
	color: white;
	font-size: 1.5rem;
	cursor: pointer;
	padding: 0;
	width: 24px;
	height: 24px;
	display: flex;
	align-items: center;
	justify-content: center;
}

/* Loading Overlay */
.loading-overlay {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(255, 255, 255, 0.9);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 9998;
}

.loading-content {
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

/* Modal Styles */
.modal-overlay {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	display: flex !important;
	align-items: center;
	justify-content: center;
	z-index: 99999 !important;
	padding: 1rem;
	backdrop-filter: blur(4px);
	visibility: visible !important;
}

.modal {
	background: white;
	border-radius: 12px;
	box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
	max-width: 500px;
	width: 100%;
	max-height: 80vh;
	overflow: hidden;
	animation: modalSlideIn 0.3s ease-out;
	position: relative;
	z-index: 100000 !important;
	display: block !important;
	visibility: visible !important;
}

@keyframes modalSlideIn {
	from {
		opacity: 0;
		transform: translateY(-20px) scale(0.95);
	}
	to {
		opacity: 1;
		transform: translateY(0) scale(1);
	}
}

.modal-header {
	padding: 1.5rem;
	border-bottom: 1px solid #e1e5e9;
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.modal-header h3 {
	margin: 0;
	color: #333;
	font-size: 1.25rem;
}

.modal-close {
	background: none;
	border: none;
	font-size: 1.5rem;
	color: #6c757d;
	cursor: pointer;
	padding: 0;
	width: 24px;
	height: 24px;
	display: flex;
	align-items: center;
	justify-content: center;
}

.modal-content {
	padding: 1.5rem;
	max-height: 60vh;
	overflow-y: auto;
}

/* Search Section */
.search-section {
	display: flex;
	flex-direction: column;
	gap: 1rem;
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
	border-radius: 8px;
	font-size: 0.95rem;
	outline: none;
	transition: all 0.2s ease;
}

.search-input input:focus {
	border-color: #667eea;
	box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

/* Search Loading */
.search-loading {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	color: #6c757d;
	padding: 1rem;
	text-align: center;
	justify-content: center;
}

.spinner-small {
	width: 16px;
	height: 16px;
	border: 2px solid #f3f3f3;
	border-top: 2px solid #667eea;
	border-radius: 50%;
	animation: spin 1s linear infinite;
}

/* Search Results */
.search-results {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	max-height: 300px;
	overflow-y: auto;
}

.search-result-item {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: 0.75rem;
	border-radius: 8px;
	cursor: pointer;
	transition: all 0.2s ease;
	border: 1px solid transparent;
}

.search-result-item:hover {
	background: #f8f9fa;
	border-color: #e1e5e9;
}

.search-result-item.disabled {
	opacity: 0.5;
	cursor: not-allowed;
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

.user-info {
	flex: 1;
}

.username {
	font-weight: 600;
	color: #333;
	margin-bottom: 0.25rem;
}

.user-action {
	font-size: 0.85rem;
	color: #6c757d;
}

/* No Results */
.no-results {
	text-align: center;
	color: #6c757d;
	padding: 2rem;
}

/* Group Creation */
.group-creation {
	display: flex;
	flex-direction: column;
	gap: 1.5rem;
}

.form-group {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.form-group label {
	font-weight: 600;
	color: #333;
	font-size: 0.95rem;
}

.form-group input {
	padding: 0.75rem;
	border: 1px solid #e1e5e9;
	border-radius: 8px;
	font-size: 0.95rem;
	outline: none;
	transition: all 0.2s ease;
}

.form-group input:focus {
	border-color: #667eea;
	box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

/* Selected Members */
.selected-members {
	display: flex;
	flex-wrap: wrap;
	gap: 0.5rem;
	margin-top: 0.5rem;
}

.selected-member {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	background: #667eea;
	color: white;
	padding: 0.5rem 0.75rem;
	border-radius: 20px;
	font-size: 0.85rem;
}

.remove-member {
	background: none;
	border: none;
	color: white;
	cursor: pointer;
	font-size: 1.2rem;
	padding: 0;
	width: 20px;
	height: 20px;
	display: flex;
	align-items: center;
	justify-content: center;
}

/* Modal Actions */
.modal-actions {
	display: flex;
	gap: 0.75rem;
	justify-content: flex-end;
	margin-top: 1.5rem;
	padding-top: 1.5rem;
	border-top: 1px solid #e1e5e9;
}

.btn {
	padding: 0.75rem 1.5rem;
	border: none;
	border-radius: 8px;
	font-size: 0.95rem;
	font-weight: 600;
	cursor: pointer;
	transition: all 0.2s ease;
}

.btn-secondary {
	background: #6c757d;
	color: white;
}

.btn-secondary:hover {
	background: #5a6268;
}

.btn-primary {
	background: #667eea;
	color: white;
}

.btn-primary:hover:not(:disabled) {
	background: #5a6fd8;
}

.btn-primary:disabled {
	background: #6c757d;
	cursor: not-allowed;
}

/* Mobile Modal Adjustments */
@media (max-width: 768px) {
	.modal {
		margin: 0;
		border-radius: 0;
		max-height: 100vh;
	}
	
	.modal-content {
		max-height: calc(100vh - 120px);
	}
}

/* Spinner Animation */
@keyframes spin {
	0% { transform: rotate(0deg); }
	100% { transform: rotate(360deg); }
}
</style>
