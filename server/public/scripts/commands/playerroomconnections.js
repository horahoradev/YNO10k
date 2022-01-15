let shouldPrintPlayerRoomConnections = false;

function PlayerRoomConnectionMessagesSwitchCommand(args) {
	shouldPrintPlayerRoomConnections = !shouldPrintPlayerRoomConnections;
	PrintChatInfo("turned " + (shouldPrintPlayerRoomConnections ? "on" : "off"), "Players connection info");
	return true;
}
