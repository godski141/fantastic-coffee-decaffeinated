import axios from "axios";

const instance = axios.create({
	baseURL: __API_URL__,
	timeout: 1000 * 5
});

// Interceptor per aggiungere automaticamente l'header Authorization
instance.interceptors.request.use(
	(config) => {
		const userId = localStorage.getItem('userId');
		if (userId) {
			config.headers.Authorization = `Bearer ${userId}`;
		}
		return config;
	},
	(error) => {
		return Promise.reject(error);
	}
);

// ===========================================
// FUNZIONI DI AUTENTICAZIONE
// ===========================================

/**
 * Effettua il login/registrazione di un utente
 * @param {string} username - Nome utente (3-50 caratteri)
 * @returns {Promise} - Risposta con user_id
 */
export const doLogin = async (username) => {
	const response = await instance.post('/session', { username });
	return response.data;
};

// ===========================================
// FUNZIONI PER CONVERSAZIONI
// ===========================================

/**
 * Recupera tutte le conversazioni dell'utente
 * @returns {Promise} - Lista delle conversazioni
 */
export const getMyConversations = async () => {
	const response = await instance.get('/conversations');
	return response.data;
};

/**
 * Recupera i dettagli di una conversazione specifica
 * @param {string} conversationId - ID della conversazione
 * @returns {Promise} - Dettagli della conversazione
 */
export const getConversation = async (conversationId) => {
	const response = await instance.get(`/conversations/get-details/${conversationId}`);
	return response.data;
};

/**
 * Crea una nuova conversazione diretta con un utente
 * @param {string} username - Nome utente con cui iniziare la conversazione
 * @returns {Promise} - ID della conversazione creata
 */
export const startConversation = async (username) => {
	const response = await instance.post('/conversations/start-conversation', { username });
	return response.data;
};

/**
 * Elimina una conversazione
 * @param {string} conversationId - ID della conversazione da eliminare
 * @returns {Promise} - Risposta vuota
 */
export const deleteConversation = async (conversationId) => {
	const response = await instance.delete(`/conversations/delete/${conversationId}`);
	return response.data;
};

/**
 * Recupera i membri di una conversazione
 * @param {string} conversationId - ID della conversazione
 * @returns {Promise} - Lista dei membri
 */
export const getConversationMembers = async (conversationId) => {
	const response = await instance.get(`/conversations/get-members/${conversationId}`);
	return response.data;
};

// ===========================================
// FUNZIONI PER MESSAGGI
// ===========================================

/**
 * Recupera tutti i messaggi di una conversazione
 * @param {string} conversationId - ID della conversazione
 * @returns {Promise} - Lista dei messaggi
 */
export const getMessages = async (conversationId) => {
	const response = await instance.get(`/conversations/messages/${conversationId}`);
	return response.data;
};

/**
 * Invia un messaggio in una conversazione
 * @param {string} conversationId - ID della conversazione
 * @param {string} content - Contenuto del messaggio
 * @returns {Promise} - Messaggio inviato
 */
export const sendMessage = async (conversationId, content) => {
	const response = await instance.post(`/conversations/send-message/${conversationId}`, { content });
	return response.data;
};

/**
 * Elimina un messaggio
 * @param {string} conversationId - ID della conversazione
 * @param {string} messageId - ID del messaggio da eliminare
 * @returns {Promise} - Risposta vuota
 */
export const deleteMessage = async (conversationId, messageId) => {
	const response = await instance.delete(`/conversations/delete-message/${conversationId}/message/${messageId}`);
	return response.data;
};

/**
 * Inoltra un messaggio a un'altra conversazione
 * @param {string} conversationId - ID della conversazione di origine
 * @param {string} messageId - ID del messaggio da inoltrare
 * @param {string} targetConversationId - ID della conversazione di destinazione
 * @returns {Promise} - Messaggio inoltrato
 */
export const forwardMessage = async (conversationId, messageId, targetConversationId) => {
	const response = await instance.post(`/conversations/forward-message/${conversationId}/messages/${messageId}`, {
		conversation_id: targetConversationId
	});
	return response.data;
};

/**
 * Risponde a un messaggio
 * @param {string} conversationId - ID della conversazione
 * @param {string} messageId - ID del messaggio a cui rispondere
 * @param {string} content - Contenuto della risposta
 * @returns {Promise} - Messaggio di risposta
 */
export const replyMessage = async (conversationId, messageId, content) => {
	const response = await instance.post(`/conversations/reply-message/${conversationId}/message/${messageId}`, { content });
	return response.data;
};

/**
 * Aggiunge una reazione emoji a un messaggio
 * @param {string} conversationId - ID della conversazione
 * @param {string} messageId - ID del messaggio
 * @param {string} emoji - Emoji da aggiungere
 * @returns {Promise} - Risposta vuota
 */
export const addReaction = async (conversationId, messageId, emoji) => {
	const response = await instance.post(`/conversations/react/${conversationId}/messages/${messageId}`, { emoji });
	return response.data;
};

/**
 * Rimuove una reazione emoji da un messaggio
 * @param {string} conversationId - ID della conversazione
 * @param {string} messageId - ID del messaggio
 * @returns {Promise} - Risposta vuota
 */
export const removeReaction = async (conversationId, messageId) => {
	const response = await instance.delete(`/conversations/delete-react/${conversationId}/messages/${messageId}`);
	return response.data;
};

// ===========================================
// FUNZIONI PER GRUPPI
// ===========================================

/**
 * Crea un nuovo gruppo
 * @param {string} name - Nome del gruppo
 * @param {string[]} members - Lista dei nomi utente dei membri
 * @returns {Promise} - ID del gruppo creato
 */
export const createGroup = async (name, members) => {
	const response = await instance.post('/conversations/create-group', { name, members });
	return response.data;
};

/**
 * Cambia il nome di un gruppo
 * @param {string} conversationId - ID del gruppo
 * @param {string} newName - Nuovo nome del gruppo
 * @returns {Promise} - Risposta vuota
 */
export const renameGroup = async (conversationId, newName) => {
	const response = await instance.patch(`/conversations/group/change-name/${conversationId}`, { name: newName });
	return response.data;
};

/**
 * Aggiunge un utente a un gruppo
 * @param {string} conversationId - ID del gruppo
 * @param {string} username - Nome utente da aggiungere
 * @returns {Promise} - Risposta vuota
 */
export const addToGroup = async (conversationId, username) => {
	const response = await instance.post(`/conversations/group/add/${conversationId}`, { username });
	return response.data;
};

/**
 * Recupera i membri di un gruppo
 * @param {string} conversationId - ID del gruppo
 * @returns {Promise} - Lista dei nomi utente dei membri
 */
export const getGroupMembers = async (conversationId) => {
	const response = await instance.get(`/conversations/group/members/${conversationId}`);
	return response.data;
};

/**
 * Esce da un gruppo
 * @param {string} conversationId - ID del gruppo
 * @returns {Promise} - Risposta vuota
 */
export const leaveGroup = async (conversationId) => {
	const response = await instance.delete(`/conversations/group/leave/${conversationId}`);
	return response.data;
};

/**
 * Cambia la foto di un gruppo
 * @param {string} conversationId - ID del gruppo
 * @param {string} photoBase64 - Foto in formato base64
 * @returns {Promise} - Risposta vuota
 */
export const updateGroupPhoto = async (conversationId, photoBase64) => {
	const response = await instance.patch(`/conversations/group/change-photo/${conversationId}`, { photo: photoBase64 });
	return response.data;
};

/**
 * Recupera la foto di un gruppo
 * @param {string} conversationId - ID del gruppo
 * @returns {Promise} - URL della foto
 */
export const getGroupPhoto = async (conversationId) => {
	const response = await instance.get(`/conversations/group/get-photo/${conversationId}`);
	return response.data;
};

// ===========================================
// FUNZIONI PER UTENTI
// ===========================================

/**
 * Cambia il nome utente
 * @param {string} newName - Nuovo nome utente
 * @returns {Promise} - Risposta vuota
 */
export const changeUsername = async (newName) => {
	const response = await instance.patch('/users/modify-username', { new_name: newName });
	return response.data;
};

/**
 * Recupera la foto di un utente
 * @param {string} userId - ID dell'utente
 * @returns {Promise} - URL della foto
 */
export const getUserPhoto = async (userId) => {
	const response = await instance.get(`/users/get-photo/${userId}`);
	return response.data;
};

/**
 * Cambia la foto profilo dell'utente
 * @param {string} photoBase64 - Foto in formato base64
 * @returns {Promise} - Risposta vuota
 */
export const updateUserPhoto = async (photoBase64) => {
	const response = await instance.patch('/users/update-photo', { photo: photoBase64 });
	return response.data;
};

/**
 * Cerca utenti per nome
 * @param {string} query - Query di ricerca
 * @returns {Promise} - Lista dei nomi utente trovati
 */
export const searchUsers = async (query) => {
	const response = await instance.get(`/user/search?username=${encodeURIComponent(query)}`);
	return response.data;
};

// ===========================================
// FUNZIONI DI UTILITÀ
// ===========================================

/**
 * Verifica se l'utente è autenticato
 * @returns {boolean} - True se autenticato
 */
export const isAuthenticated = () => {
	return !!localStorage.getItem('userId');
};

/**
 * Effettua il logout
 */
export const logout = () => {
	localStorage.removeItem('userId');
	localStorage.removeItem('username');
	localStorage.removeItem('userPhoto');
};

/**
 * Salva i dati utente nel localStorage
 * @param {string} userId - ID utente
 * @param {string} username - Nome utente
 * @param {string} userPhoto - Foto utente (opzionale)
 */
export const saveUserData = (userId, username, userPhoto = null) => {
	localStorage.setItem('userId', userId);
	localStorage.setItem('username', username);
	if (userPhoto) {
		localStorage.setItem('userPhoto', userPhoto);
	}
};

/**
 * Recupera i dati utente dal localStorage
 * @returns {Object} - Dati utente
 */
export const getUserData = () => {
	return {
		userId: localStorage.getItem('userId'),
		username: localStorage.getItem('username'),
		userPhoto: localStorage.getItem('userPhoto')
	};
};

export default instance;
