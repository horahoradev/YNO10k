ENV.SDL_EMSCRIPTEN_KEYBOARD_ELEMENT = "#canvas";

let chatMessagesContainer = document.getElementById("messages");
let chatInputContainer = document.getElementById("chatInputContainer");
let enterChatForm = document.getElementById('enterChatForm');
let nameInput = document.getElementById('nameInput');
let tripInput = document.getElementById('tripInput');
let chatInput = document.getElementById("chatInput")

let profilepacket;

let globalMessagesToggle = document.getElementById("globalMessagesToggle");
let globalChatDisplayed = true;

globalMessagesToggle.onclick = function(){
	globalChatDisplayed = !globalChatDisplayed;
	document.querySelector(':root').style.setProperty('--global-chat-display', globalChatDisplayed ? 'block' : 'none');
}

// split into two functions since game client can now call this
function SendMessageString(textStr) {
  if (textStr === "") {
    return;
  }

  if(textStr[0] == '/') {
	  ExecuteCommand(textStr);
	  return;
  }

  if(textStr[0] == '!') {
  	textStr = textStr.substr(1);
	YNOnline.Network.globalChat.SendMessage(JSON.stringify({text: textStr}));
  } else if(YNOnline.Network.localChat){
	YNOnline.Network.localChat.SendMessage(JSON.stringify({text: textStr}));
  } else {
	  PrintChatInfo("You're not connected to any room\nYou can use global chat with '!' at the begining of a message", "Client");
  }
}

// called by HTML typebox
function SendMessage() {
	SendMessageString(chatInput.value);
	chatInput.value = "";
}

function NewChatMessage(message, source) {
	let messageContainer = document.createElement("div");
	let messageTextContainer = document.createElement("div");
	let messageProfileContainer = document.createElement("div");
	let messageProfileNameContainer = document.createElement("div");
	let messageProfileTripContainer = document.createElement("div");
	let messageProfileSourceContainer = document.createElement("div");
	messageContainer.classList.add("MessageContainer", source);
	messageTextContainer.className = "MessageTextContainer";
	messageProfileContainer.className = "MessageProfileContainer";
	messageProfileNameContainer.className = "MessageProfileNameContainer";
	messageProfileTripContainer.className = "MessageProfileTripContainer";
	messageProfileSourceContainer.className = "MessageProfileSourceContainer";
	
	messageProfileNameContainer.innerText = message.name;
	messageProfileTripContainer.innerText = message.trip;
	messageProfileSourceContainer.innerText = source;
	messageTextContainer.innerText = message.text;

	messageProfileContainer.append(messageProfileSourceContainer);
	messageProfileContainer.append(messageProfileNameContainer);
	messageProfileContainer.append(messageProfileTripContainer);
	messageContainer.append(messageProfileContainer);
	messageContainer.appendChild(messageTextContainer);
	chatMessagesContainer.appendChild(messageContainer);
	onNewChatEntry();

	// let game know about incoming messages
	let inMsgName = Module.allocate(Module.intArrayFromString(message.name), Module.ALLOC_NORMAL);
	let inMsgTrip = Module.allocate(Module.intArrayFromString(message.trip), Module.ALLOC_NORMAL);
	let inMsgText = Module.allocate(Module.intArrayFromString(message.text), Module.ALLOC_NORMAL);
	let inMsgSrc = Module.allocate(Module.intArrayFromString(source), Module.ALLOC_NORMAL);
  Module._gotMessage(inMsgName, inMsgTrip, inMsgText, inMsgSrc);
	Module._free(inMsgName);
	Module._free(inMsgTrip);
	Module._free(inMsgText);
	Module._free(inMsgSrc);
}

function PrintChatInfo(text, source) {
	let infoContainer = document.createElement("div");
	let infoTextContainer = document.createElement("div");
	let infoSourceContainer = document.createElement("div");
	infoContainer.className = "InfoContainer";
	infoSourceContainer.className = "InfoSourceContainer";
	infoTextContainer.className = "InfoTextContainer";
	infoTextContainer.innerText = text;
	infoSourceContainer.innerText = source + ":";
	infoContainer.appendChild(infoSourceContainer);
	infoContainer.appendChild(infoTextContainer);
	chatMessagesContainer.appendChild(infoContainer);
	onNewChatEntry();

	// let game know about incoming info
	let inMsgSource = Module.allocate(Module.intArrayFromString(source), Module.ALLOC_NORMAL);
	let inMsgText = Module.allocate(Module.intArrayFromString(text), Module.ALLOC_NORMAL);
  Module._gotChatInfo(inMsgSource, inMsgText);
	Module._free(inMsgSource);
	Module._free(inMsgText);
}

function onNewChatEntry() {
	let shouldScroll = (chatMessagesContainer.scrollHeight - chatMessagesContainer.scrollTop - chatMessagesContainer.clientHeight) <= 100;
	if(shouldScroll) {
		chatMessagesContainer.scrollTop = chatMessagesContainer.scrollHeight;
	}
}

function Chat (address, chatname, isglobal) 
{
	let preConnectionMessageQueue = [];
	let socket = new WebSocket(address);
	let chatType = "L";
	this.isglobal = isglobal;
	let self = this;
	let shouldBeClosed = false;

	if(isglobal)
		chatType = "G";
	
	socket.onopen = function(e) {
		socket.send(gameName + "chat");
		socket.send(chatname);

		for(m of preConnectionMessageQueue) {
			socket.send(m);
		}
		if(profilepacket)
			socket.send(profilepacket);
		delete preConnectionMessageQueue;
	}

	socket.onmessage = function(e) {
		let data = JSON.parse(e.data);
		switch(data.type) {
		case "userMessage":
			NewChatMessage(data, chatType);
		break;
		case "userConnect":
			if(self.isglobal)
				PrintChatInfo(" " + data.name + " joined chat.", "Server");
			else if(shouldPrintPlayerRoomConnections)
				PrintChatInfo(" " + data.name + " joined this room.", "Server");
		break;
		case "roomDisconnect":
			if(self.isglobal)
				PrintChatInfo(" " + data.name + " left.", "Server");
			else if(shouldPrintPlayerRoomConnections)
 				PrintChatInfo(" " + data.name + " left this room.", "Server");
		break;
		case "serverInfo":
			PrintChatInfo(data.text, "Server")
		break;
		case "playersCount":
		if(data.count > 1) {
			if(self.isglobal)
				PrintChatInfo(" " + data.count + " players online in total.", "Server");
			else
				PrintChatInfo(" " + data.count + " players in this room.", "Server");
		}
		break;
		case "ping":
			self.SendMessage("{ \"type\": \"pong\" }");
		break;
		}
	}

	socket.onclose = function() {
		if(!shouldBeClosed) {
			setTimeout( function() {
			let s = new WebSocket(address);
			s.onopen = socket.onopen;
			s.onmessage = socket.onmessage;
			s.onclose = socket.onclose;
			socket = s;
			}, 5000);
		}
	}

	this.SendMessage = function(message) {
		if(socket.readyState === WebSocket.CONNECTING) {
			preConnectionMessageQueue.push(message);
		} else {
			socket.send(message);
		}
	}

	this.Close = function() {
		shouldBeClosed = true;
		socket.close();
	}
}

function randomTripcode(len) {
	let t = "";
	let c = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRTSUVWXYZ0123456789";
	for(let i = 0; i < len; i++)
		t += c[parseInt(Math.random() * c.length)];
	return t;
}

// called by both HTML chat and in-game chat. Each one has its own validation process before calling this.
function SendProfileInfo(iName, iTrip) {
	// no trip specified, generate random.
	if(iTrip === "") iTrip = randomTripcode(16);

	// send to server
	let data = {name: iName, trip: iTrip}
	profilepacket = JSON.stringify(data);
	YNOnline.Network.globalChat.SendMessage(profilepacket);
	// update info in-game
	let name = Module.allocate(Module.intArrayFromString(iName), Module.ALLOC_NORMAL);
  Module._ChangeName(name);
	Module._free(name);
	// connect
	ConnectToLocalChat(GetRoomID());

	// Change chat interface to allow for message input. Necessary whether profile info was sent from HTML or in-game chat.
	// TO-DO: sending profile info from HTML chat instead does not update name input interface in-game.
	// Shouldn't be a problem if HTML chat will be completely replaced by in-game one.
	document.getElementById("enterChatContainer").style.display = "none";
	chatInputContainer.style.display = "block";
	chatInput.disabled = false;

	// TO-DO: chat preferences only load in HTML profile input. Not in-game.
	nameInput.value = iName; // Sets values of input fields because saveChatConfig() reads from them.
	tripInput.value = iTrip; // This is done so it can save preferences even if profile info is sent from in-game.
	saveChatConfig(); // <-- reads from HTML fields.
	nameInput.value = "";
	tripInput.value = "";
}

function TrySendProfileInfo() {
	// Separate validation (for HTML chat) from actual sending done in SendProfileInfo().
	// We don't want validation to fail here on the JS side after game client (in-game chat) has already sent and hidden its name input box,
	// 		so game client will do validation of its own before calling SendProfileInfo().

	if(nameInput.value === "") return; // additional restrictions enforced by HTML (max length and latin alphanumeric characters)
	// valid, send.
	SendProfileInfo(nameInput.value, tripInput.value);
}

function ConnectToLocalChat(room) {
	if(YNOnline.Network.localChat)
		YNOnline.Network.localChat.Close();
	YNOnline.Network.localChat = new Chat(WSAddress, "chat"+room);
	if(profilepacket)
		YNOnline.Network.localChat.SendMessage(profilepacket);
}

YNOnline.Network.globalChat = null;
YNOnline.Network.localChat = null;

function initChat() {
	// loads user config only after module has been initialized,
	// so we can communicate with game (to send user preferences)
	loadOrInitConfig();

	moduleInitialized = true;

	// open global chat only after module has been initialized,
	// so it can send info messages to the in-game chat
	YNOnline.Network.globalChat = new Chat(WSAddress, "gchat", true);

	PrintChatInfo("Type /help to see list of slash commands.\nUse '!' at the bebining of a message to send it to global chat.", "Info");
}

window.onresize = function(event) {
    if(document.documentElement.clientWidth < 1190) {
		document.getElementById("chatboxContainer").style.width = "100%";
	} else {
		document.getElementById("chatboxContainer").style.width = "calc(100% - 66%)";
	}
};

window.onresize();

/*
		============================
		============================
		HTML chat helper integration
		============================
		============================
*/

var moduleInitialized = false; // to know if we can call Module functions
var chatHelper = document.getElementById("chat_input_helper");
var gameCanvas = document.getElementById("canvas");

/*
	Clone HTML chat input events into the game canvas.
*/

function rerouteEvent(event) {
	var newEvent = new event.constructor(event.type, event);
	gameCanvas.dispatchEvent(newEvent);
}
chatHelper.addEventListener("contextmenu", rerouteEvent);
chatHelper.addEventListener("keydown", rerouteEvent);
chatHelper.addEventListener("keypress", rerouteEvent);
chatHelper.addEventListener("keyup", rerouteEvent);
chatHelper.addEventListener("mousedown", rerouteEvent);
chatHelper.addEventListener("mouseenter", rerouteEvent);
chatHelper.addEventListener("mouseleave", rerouteEvent);
chatHelper.addEventListener("mousemove", rerouteEvent);
chatHelper.addEventListener("touchcancel", rerouteEvent);
chatHelper.addEventListener("touchend", rerouteEvent);
chatHelper.addEventListener("touchmove", rerouteEvent);
chatHelper.addEventListener("touchstart", rerouteEvent);
chatHelper.addEventListener("webglcontextlost", rerouteEvent);
chatHelper.addEventListener("wheel", rerouteEvent);

/*
	Tie HTML chat helper state to the in-game chatbox's state.
	Chat helper should be focused if and only if game is tabbed into chat.
*/

// called by game to make HTML chat mirror its focused state
function setChatFocus(focused) {
	if(focused) {
		chatHelper.focus();
	} else {
		gameCanvas.focus();
	}
}
// prevent HTML chat input from focusing when in-game chat is closed
chatHelper.addEventListener("focus", function() {
	if(!moduleInitialized) {
		chatHelper.blur();
	} else {
		var gameChatOpen = Module._isChatOpen();
		if(gameChatOpen == 0) {
			gameCanvas.focus();
		}
	}
});
// prevent HTML chat input from unfocusing when in-game chat is open
gameCanvas.addEventListener("focus", function() {
	if(!moduleInitialized) return;
	var gameChatOpen = Module._isChatOpen();
	if(gameChatOpen == 1) {
		chatHelper.focus();
	}
});
// prevent tabbing out from HTML chat input to other page elements
function preventTab(event) {
	if(event.keyCode === 9) {
		event.preventDefault();
  }
}
chatHelper.addEventListener("keydown", preventTab);

/*
	Called by game to set type box's state
*/

function setTypeText(text) {
	chatHelper.value = text;
	typeUpdated();
}

function setTypeMaxChars(c) {
	chatHelper.maxLength = c;
}

/*
	Feed HTML chat helper data into the game.
*/

var previousTypeText = ""; // store previously commited text so helper only calls game to redraw if it changed.

function typeUpdated() {
	if(!moduleInitialized) {
		chatHelper.value = "";
	} else {
		setTimeout(function() { // We need a small time delay provided by setTimeout to get the correct caret position after moving it.
														// Without it, selectionStart and selectionEnd would have the caret's previous position.
	    var cTail = chatHelper.selectionStart;
	    var cHead = chatHelper.selectionEnd;
	    if(chatHelper.selectionDirection === "backward") {
	    	var tmp = cTail;
	    	cTail = cHead;
	    	cHead = tmp;
	    }
	    if(previousTypeText === chatHelper.value) { // text is the same, only move caret
	    	Module._updateTypeDisplayCaret(cTail, cHead);
	    } else { // redraw text and move caret
	    	previousTypeText = chatHelper.value;
	    	let text = Module.allocate(Module.intArrayFromString(chatHelper.value), Module.ALLOC_NORMAL);
				Module._updateTypeDisplayText(text, cTail, cHead);
				Module._free(text);
	    }
	  }, 5);
	}
}
function trySend(event) {
	if(!moduleInitialized) return;
	if(event.which === 13) {
		let text = Module.allocate(Module.intArrayFromString(chatHelper.value), Module.ALLOC_NORMAL);
		Module._trySendChat(text);
		Module._free(text);
	}
}
chatHelper.addEventListener("input", typeUpdated);
chatHelper.addEventListener("keydown", typeUpdated);
chatHelper.addEventListener("keydown", trySend);