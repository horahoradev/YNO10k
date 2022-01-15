// TODO s
function IgnoreChatCommand(args) {
	if(args.length == 2) {
	YNOnline.Network.globalChat.SendMessage(JSON.stringify({
		ignorechat: {uuid: args[1]}
		}));
		return true;
	} else {
		return false;
	}
}

function IgnoreGameCommand(args) {
	if(args.length == 2) {
	YNOnline.Network.globalChat.SendMessage(JSON.stringify({
		ignoregame: {uuid: args[1]}
		}));
		return true;
	} else {
		return false;
	}
}
