
// TODO s
function PardonChatCommand(args) {
	if(args.length == 2) {
	YNOnline.Network.globalChat.SendMessage(JSON.stringify({
		pardonchat: {uuid: args[1]}
		}));
		return true;
	} else {
		return false;
	}
}

function PardonGameCommand(args) {
	if(args.length == 2) {
	YNOnline.Network.globalChat.SendMessage(JSON.stringify({
		pardongame: {uuid: args[1]}
		}));
		return true;
	} else {
		return false;
	}
}
