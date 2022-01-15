
function GetUUIDCommand(args) {
	if(args.length == 1) {
		YNOnline.Network.globalChat.SendMessage(JSON.stringify({getuuid: "*", room: GetRoomID()}));
	} else if(args.length == 2) {
		YNOnline.Network.globalChat.SendMessage(JSON.stringify({getuuid: args[1]}));
	} else {
		return false;
	}
	return true;
}