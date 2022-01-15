
let spritelist = {
	monoe: {sheet: "000000000054", id: 0}
}

let helpData = {

	help: {
		tip: "display this list\n/help <command> usage of command",
		help: 
		"/help\n/help <command> to see usage of command"
		},

	spriteset: {
		tip: "set your sprite",
		help: 
		"/spriteset <sprite_name> (see list of sprite names with /spritelist)\n/spriteset <sheet> <id> (number from 0 to 7)"
	},

	spritelist: {
		tip: "displays list of sprite names",
		help: "/spritelist"
	},

	ignorechat: {
		tip: "ignore player chat messages by uuid, also ignores player by ip on server side",
		help: "/ignorechat <uuid> you can display list of player uuids with /getuuid or /getuuid <tripcode>"
	},

	ignoregame: {
		tip: "ignore player in game by uuid, also ignores player by ip on server side",
		help: "/ignoregame <uuid> you can display list of players uuid with /getuuid or /getuuid <tripcode>"
	},

	getuuid: {
		tip: "displays list of uuids of players in your room or uuid of a player by tripcode",
		help: "/getuuid for list of uuids\n/getuuid <tripcode> for specific user uuid"
	},

	pvol: {
		tip: "sets other players sound effects volume (50% by default)",
		help: "/pvol <volume> (number from 0 to 100)"
	},

	pardonchat: {
		tip: "Lets you unignore players in chat",
		help: "/pardonchat <uuid>"
	},

	pardongame: {
		tip: "Lets you unignore players in game",
		help: "/pardongame <uuid>"
	},

	rcm: {
		tip: "Switch room connection messages on/off",
		help: "/rcm"
	},

	pcm: {
		tip: "Switch messages about players connecting/disconnecting to current room on/off",
		help: "/pcm"
	},

	spritefav: {
		tip: "Make shortcut for sprite sheet and index to use with /spriteset",
		help: "/spritefav <sheet> <index> <name>"
	},

	spriteunfav: {
		tip: "Remove shortcut for sprite",
		help: "/spriteunfav <name>"
	}
}

function HelpCommand(args) {
	if(args.length == 1) {
		let helpkeys = Object.keys(helpData);
		for(let k of helpkeys) {
			PrintChatInfo("/" + k + "\n"+helpData[k].tip, "Help");
		}
	} else if(args.length == 2) {
		if(helpData[args[1]]) {
			PrintChatInfo(helpData[args[1]].help, "Help")
		}
		else {
			PrintChatInfo("Unknown command", "Help")
		}
	} else {
		return false;
	}
	return true;
}