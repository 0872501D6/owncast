class Message {
	constructor(model) {
		this.author = model.author
		this.body = model.body
		this.image = "https://robohash.org/" + model.author
		this.id = model.id
	}

	addNewlines(str) {
		return str.replace(/(?:\r\n|\r|\n)/g, '<br />');

	}
	formatText() {
		var linked = autoLink(this.body, { embed: true });
		return this.addNewlines(linked);
	}

	toModel() {
		return {
			author: this.author(),
			body: this.body(),
			image: this.image(),
			id: this.id
		}
	}
}


// convert newlines to <br>s

class Messaging {
	constructor() {
		this.chatDisplayed = false;
		this.username = "";
		this.avatarSource = "https://robohash.org/";

		this.messageCharCount = 0;
		this.maxMessageLength = 500;
		this.maxMessageBuffer = 20;

		this.keyUsername = "owncast_username";
		this.keyChatDisplayed = "owncast_chat";

		this.tagChatToggle = document.querySelector("#chat-toggle");

		this.tagUserInfoDisplay = document.querySelector("#user-info-display");
		this.tagUserInfoChanger = document.querySelector("#user-info-change");

		this.tagUsernameDisplay = document.querySelector("#username-display");
		this.imgUsernameAvatar = document.querySelector("#username-avatar");

		this.inputMessageAuthor = document.querySelector("#self-message-author");

		this.tagMessageFormWarning = document.querySelector("#message-form-warning");
		
		this.tagAppContainer = document.querySelector("#app-container");

		this.inputChangeUserName = document.querySelector("#username-change-input"); 
		this.btnUpdateUserName = document.querySelector("#button-update-username"); 
		this.btnCancelUpdateUsername = document.querySelector("#button-cancel-change"); 

		this.formMessageInput = document.querySelector("#inputBody");
	}
	init() {
		this.tagChatToggle.addEventListener("click", this.handleChatToggle);
		this.tagUsernameDisplay.addEventListener("click", this.handleShowChangeNameForm);
		
		this.btnUpdateUserName.addEventListener("click", this.handleUpdateUsername);
		this.btnCancelUpdateUsername.addEventListener("click", this.handleHideChangeNameForm);
		
		this.inputChangeUserName.addEventListener("keydown", this.handleUsernameKeydown);
		this.formMessageInput.addEventListener("keydown", this.handleMessageInputKeydown);
		this.initLocalStates();

		var vueMessageForm = new Vue({
			el: "#message-form",
			data: {
				message: {
					author: "",//localStorage.author || "Viewer" + (Math.floor(Math.random() * 42) + 1),
					body: ""
				}
			},
			methods: {
				submitChatForm: function (e) {
					const message = new Message(this.message);
					message.id = uuidv4();
					localStorage.author = message.author;
					const messageJSON = JSON.stringify(message);
					window.ws.send(messageJSON);
					e.preventDefault();
	
					this.message.body = "";
				}
			}
		});

		
	}

	initLocalStates() {
		this.username = getLocalStorage(this.keyUsername) || "User" + (Math.floor(Math.random() * 42) + 1);
		this.updateUsernameFields(this.username);

		this.chatDisplayed = getLocalStorage(this.keyChatDisplayed) || false;
		this.displayChat();
	}
	updateUsernameFields(username) {
		this.tagUsernameDisplay.innerText = username;
		this.inputChangeUserName.value = username;
		this.inputMessageAuthor.value = username;
		this.imgUsernameAvatar.src = this.avatarSource + username;
	}
	displayChat() {
		this.tagAppContainer.className = this.chatDisplayed ?  "flex" : "flex no-chat";
	}

	handleChatToggle = () => {
		this.chatDisplayed = !this.chatDisplayed;

		setLocalStorage(this.keyChatDisplayed, this.chatDisplayed);
		this.displayChat();
	}
	handleShowChangeNameForm = () => {
		this.tagUserInfoDisplay.style.display = "none";
		this.tagUserInfoChanger.style.display = "flex";
	}
	handleHideChangeNameForm = () => {
		this.tagUserInfoDisplay.style.display = "flex";
		this.tagUserInfoChanger.style.display = "none";
	}
	handleUpdateUsername = () => {
		var newValue = this.inputChangeUserName.value;
		newValue = newValue.trim();
		// do other string cleanup?

		if (newValue) {
			this.userName = newValue;
			this.updateUsernameFields(newValue);
			// this.inputChangeUserName.value = newValue;
			// this.inputMessageAuthor.value = newValue;
			// this.tagUsernameDisplay.innerText = newValue;
			// this.imgUsernameAvatar.src = this.avatarSource + newValue;
			setLocalStorage(this.keyUsername, newValue);
		}
		this.handleHideChangeNameForm();
	}

	handleUsernameKeydown = event => {
		if (event.keyCode === 13) { // enter
			this.handleUpdateUsername();
		} else if (event.keyCode === 27) { // esc
			this.handleHideChangeNameForm();
		}
	}

	handleMessageInputKeydown = event => {
		var okCodes = [37,38,39,40,16,91,18,46,8];
		var value = this.formMessageInput.value.trim();
		var numCharsLeft = this.maxMessageLength - value.length;

		if (event.keyCode === 13) { // enter
			if (!this.prepNewLine) {
				// submit()
				event.preventDefault();
				// clear out things.
				this.formMessageInput.value = "";
				this.tagMessageFormWarning.innerText = "";
				return;
			}
			this.prepNewLine = false;
		} else {
			this.prepNewLine = false;
		}
		if (event.keyCode === 16 || event.keyCode === 17) { // ctrl, shift
			this.prepNewLine = true;
		}

		if (numCharsLeft <= this.maxMessageBuffer) {
			this.tagMessageFormWarning.innerText = numCharsLeft + " chars left";
			if (numCharsLeft <= 0 && !okCodes.includes(event.keyCode)) {
				event.preventDefault();
				return;
			}
		} else {
			this.tagMessageFormWarning.innerText = "";
		}
	}



}