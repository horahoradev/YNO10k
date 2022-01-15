let shouldPrintRoomConnetionMessages = false;

function RoomConnectionMessagesSwitchCommand(args) {
	shouldPrintRoomConnetionMessages = !shouldPrintRoomConnetionMessages;
	PrintChatInfo("turned " + (shouldPrintRoomConnetionMessages ? "on" : "off"), "Your connection info");
	return true;
}
