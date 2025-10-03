#pragma semicolon 1
#pragma newdecls required

#include <sourcemod>
#include <sdktools>
#include <sdkhooks>

#define PLUGIN_VERSION "1.0.0"

public Plugin myinfo = {
    name = "Bot Mimic Player",
    author = "AI Assistant",
    description = "Makes bots mimic a player's actions when aimed at and side mouse button is held",
    version = PLUGIN_VERSION,
    url = ""
};

// Constants - Mouse buttons in CS:GO are different from standard values
// IN_USE = (1<<5) is E key, often bound to mouse4 by default
// Try different values for mouse buttons
#define MOUSE4 (1<<5)  // Usually bound to IN_USE (E key)
#define MOUSE5 (1<<6)  // Usually bound to IN_ATTACK2 (right click)
#define MIMIC_CHECK_INTERVAL 0.1

// ConVars
ConVar g_cvMimicButton;
ConVar g_cvDebugMode;
int g_iMimicButton;
bool g_bDebugMode;

// Global variables
bool g_bPlayerMimicking[MAXPLAYERS + 1]; // Tracks if a player is making a bot mimic them
int g_iMimicBot[MAXPLAYERS + 1]; // The bot that a player is controlling
Handle g_hMimicTimer[MAXPLAYERS + 1]; // Timer handle for each player
bool g_bIsRecording[MAXPLAYERS + 1]; // Tracks if we're recording a player for bot mimic

public void OnPluginStart()
{
    // Register commands
    RegConsoleCmd("sm_mimic_status", Command_MimicStatus, "Show the current mimic status");
    RegConsoleCmd("sm_mimic_debug", Command_MimicDebug, "Debug button presses");
    
    // Create ConVars
    g_cvMimicButton = CreateConVar("sm_botmimic_button", "5", "Button to use for bot mimicking (5=IN_USE/E, 6=IN_ATTACK2/Mouse2)", _, true, 1.0, true, 26.0);
    g_iMimicButton = (1 << g_cvMimicButton.IntValue);
    g_cvMimicButton.AddChangeHook(OnConVarChanged);
    
    g_cvDebugMode = CreateConVar("sm_botmimic_debug", "1", "Enable debug mode (0=off, 1=on)", _, true, 0.0, true, 1.0);
    g_bDebugMode = g_cvDebugMode.BoolValue;
    g_cvDebugMode.AddChangeHook(OnConVarChanged);
    
    // Hook events
    HookEvent("player_spawn", Event_PlayerSpawn);
    HookEvent("player_death", Event_PlayerDeath);
    HookEvent("round_start", Event_RoundStart);
    HookEvent("round_end", Event_RoundEnd);
    
    // Initialize arrays
    for (int i = 1; i <= MaxClients; i++)
    {
        g_bPlayerMimicking[i] = false;
        g_iMimicBot[i] = -1;
        g_hMimicTimer[i] = null;
        g_bIsRecording[i] = false;
    }
    
    // Enable sv_cheats if not already enabled
    ServerCommand("sv_cheats 1");
    
    // Print plugin loaded message
    PrintToServer("[Bot Mimic] Plugin loaded successfully!");
    
    // Create config file
    AutoExecConfig(true, "bot_mimic_player");
}

public void OnConVarChanged(ConVar convar, const char[] oldValue, const char[] newValue)
{
    if (convar == g_cvMimicButton)
    {
        g_iMimicButton = (1 << g_cvMimicButton.IntValue);
        PrintToServer("[Bot Mimic] Mimic button changed to %d (flag: %d)", g_cvMimicButton.IntValue, g_iMimicButton);
    }
    else if (convar == g_cvDebugMode)
    {
        g_bDebugMode = g_cvDebugMode.BoolValue;
    }
}

public void OnMapStart()
{
    // Print plugin loaded message to all players
    CreateTimer(5.0, Timer_AnnouncePlugin);
}

public Action Timer_AnnouncePlugin(Handle timer)
{
    PrintToChatAll("[Bot Mimic] Plugin loaded successfully! Aim at a bot and press the mimic button (default: E key) to control it.");
    return Plugin_Stop;
}

public void OnClientPutInServer(int client)
{
    if (!IsFakeClient(client))
    {
        g_bPlayerMimicking[client] = false;
        g_iMimicBot[client] = -1;
        g_hMimicTimer[client] = null;
        g_bIsRecording[client] = false;
        
        // Notify the player about the plugin
        CreateTimer(3.0, Timer_NotifyPlayer, client);
    }
}

public Action Timer_NotifyPlayer(Handle timer, any client)
{
    if (IsClientInGame(client) && !IsFakeClient(client))
    {
        PrintToChat(client, "[Bot Mimic] Aim at a bot and press the mimic button (default: E key) to control it.");
    }
    return Plugin_Stop;
}

public void OnClientDisconnect(int client)
{
    if (!IsFakeClient(client))
    {
        StopMimicking(client);
    }
}

public Action OnPlayerRunCmd(int client, int &buttons, int &impulse, float vel[3], float angles[3], int &weapon, int &subtype, int &cmdnum, int &tickcount, int &seed, int mouse[2])
{
    // Check if the player is pressing or releasing the mimic button
    if (!IsFakeClient(client) && IsPlayerAlive(client))
    {
        bool isPressingButton = (buttons & g_iMimicButton) != 0;
        bool wasMimicking = g_bPlayerMimicking[client];
        
        if (isPressingButton && !wasMimicking)
        {
            // Player just pressed the button, try to start mimicking
            int target = GetClientAimTarget(client, true);
            
            // Check if the target is a valid bot
            if (target != -1 && IsClientInGame(target) && IsFakeClient(target) && IsPlayerAlive(target))
            {
                PrintToChat(client, "[Bot Mimic] Starting to mimic bot %N...", target);
                StartMimicking(client, target);
            }
            else if (target != -1)
            {
                PrintToChat(client, "[Bot Mimic] You must aim at a valid bot to mimic it.");
            }
        }
        else if (!isPressingButton && wasMimicking)
        {
            // Player just released the button, stop mimicking
            StopMimicking(client);
        }
        else if (wasMimicking)
        {
            // Player is still holding the button, check if still aiming at the bot
            int target = GetClientAimTarget(client, true);
            if (target != g_iMimicBot[client])
            {
                // Player is no longer aiming at the same bot
                PrintToChat(client, "[Bot Mimic] You are no longer aiming at the bot. Stopping mimic.");
                StopMimicking(client);
            }
        }
    }
    
    return Plugin_Continue;
}

public Action Command_MimicDebug(int client, int args)
{
    if (client == 0)
    {
        ReplyToCommand(client, "This command can only be used in-game.");
        return Plugin_Handled;
    }
    
    // Start a timer to monitor button presses
    CreateTimer(0.1, Timer_DebugButtons, client, TIMER_REPEAT | TIMER_FLAG_NO_MAPCHANGE);
    ReplyToCommand(client, "[Bot Mimic Debug] Press buttons to see their values. Debug will run for 10 seconds.");
    
    // Create a timer to stop the debug after 10 seconds
    CreateTimer(10.0, Timer_StopDebug, client);
    
    return Plugin_Handled;
}

public Action Timer_DebugButtons(Handle timer, any client)
{
    if (!IsClientInGame(client))
        return Plugin_Stop;
    
    int buttons = GetClientButtons(client);
    if (buttons != 0)
    {
        PrintToChat(client, "[Bot Mimic Debug] Buttons: %d | Current mimic button flag: %d", buttons, g_iMimicButton);
        
        // Check specific buttons
        if (buttons & (1<<0)) PrintToChat(client, "- IN_ATTACK (1<<0) pressed");
        if (buttons & (1<<1)) PrintToChat(client, "- IN_JUMP (1<<1) pressed");
        if (buttons & (1<<2)) PrintToChat(client, "- IN_DUCK (1<<2) pressed");
        if (buttons & (1<<3)) PrintToChat(client, "- IN_FORWARD (1<<3) pressed");
        if (buttons & (1<<4)) PrintToChat(client, "- IN_BACK (1<<4) pressed");
        if (buttons & (1<<5)) PrintToChat(client, "- IN_USE (1<<5) pressed");
        if (buttons & (1<<6)) PrintToChat(client, "- IN_CANCEL (1<<6) pressed");
        if (buttons & (1<<7)) PrintToChat(client, "- IN_LEFT (1<<7) pressed");
        if (buttons & (1<<8)) PrintToChat(client, "- IN_RIGHT (1<<8) pressed");
        if (buttons & (1<<9)) PrintToChat(client, "- IN_MOVELEFT (1<<9) pressed");
        if (buttons & (1<<10)) PrintToChat(client, "- IN_MOVERIGHT (1<<10) pressed");
        if (buttons & (1<<11)) PrintToChat(client, "- IN_ATTACK2 (1<<11) pressed");
    }
    
    return Plugin_Continue;
}

public Action Timer_StopDebug(Handle timer, any client)
{
    PrintToChat(client, "[Bot Mimic Debug] Debug session ended.");
    return Plugin_Stop;
}

void StartMimicking(int client, int bot)
{
    // Check if already mimicking
    if (g_bPlayerMimicking[client])
    {
        // If mimicking a different bot, stop the current mimic
        if (g_iMimicBot[client] != bot)
        {
            StopMimicking(client);
        }
        else
        {
            // Already mimicking this bot, do nothing
            return;
        }
    }
    
    // Ensure sv_cheats is enabled
    ServerCommand("sv_cheats 1");
    
    // Set up mimicking
    g_bPlayerMimicking[client] = true;
    g_iMimicBot[client] = bot;
    
    // Create a timer to update the bot's mimic
    g_hMimicTimer[client] = CreateTimer(MIMIC_CHECK_INTERVAL, Timer_UpdateMimic, client, TIMER_REPEAT);
    
    // Notify the player
    PrintToChat(client, "[Bot Mimic] You are now controlling bot %N", bot);
    
    // Debug info
    PrintToServer("[Bot Mimic] Starting mimic: Player %N -> Bot %N", client, bot);
}

void StopMimicking(int client)
{
    // Check if the player is mimicking
    if (g_bPlayerMimicking[client])
    {
        // Stop the timer
        if (g_hMimicTimer[client] != null)
        {
            KillTimer(g_hMimicTimer[client]);
            g_hMimicTimer[client] = null;
        }
        
        // Stop the bot from mimicking if it exists and is valid
        int bot = g_iMimicBot[client];
        if (bot != -1 && IsClientInGame(bot) && IsFakeClient(bot))
        {
            // Notify the player
            PrintToChat(client, "[Bot Mimic] You are no longer controlling bot %N", bot);
        }
        
        // Reset the player's mimicking status
        g_bPlayerMimicking[client] = false;
        g_iMimicBot[client] = -1;
    }
}

public Action Timer_UpdateMimic(Handle timer, any client)
{
    // Check if the client is still valid and mimicking
    if (!IsClientInGame(client) || !IsPlayerAlive(client) || !g_bPlayerMimicking[client])
    {
        g_hMimicTimer[client] = null;
        return Plugin_Stop;
    }
    
    int bot = g_iMimicBot[client];
    
    // Check if the bot is still valid
    if (bot == -1 || !IsClientInGame(bot) || !IsPlayerAlive(bot) || !IsFakeClient(bot))
    {
        StopMimicking(client);
        return Plugin_Stop;
    }
    
    // Make the bot mimic the player
    char clientName[MAX_NAME_LENGTH];
    char botName[MAX_NAME_LENGTH];
    GetClientName(client, clientName, sizeof(clientName));
    GetClientName(bot, botName, sizeof(botName));
    
    // Use bot_mimic command with full names to ensure it works
    ServerCommand("bot_mimic \"%s\" \"%s\"", botName, clientName);
    
    // Debug output
    if (g_bDebugMode)
    {
        PrintToServer("[Bot Mimic Debug] Command: bot_mimic \"%s\" \"%s\"", botName, clientName);
    }
    
    return Plugin_Continue;
}

public Action Command_MimicStatus(int client, int args)
{
    if (client == 0)
    {
        ReplyToCommand(client, "This command can only be used in-game.");
        return Plugin_Handled;
    }
    
    ReplyToCommand(client, "[Bot Mimic] Current mimic button: %d (flag: %d)", g_cvMimicButton.IntValue, g_iMimicButton);
    
    if (g_bPlayerMimicking[client])
    {
        int bot = g_iMimicBot[client];
        if (bot != -1 && IsClientInGame(bot))
        {
            ReplyToCommand(client, "[Bot Mimic] You are currently controlling bot %N", bot);
        }
        else
        {
            ReplyToCommand(client, "[Bot Mimic] You are controlling an invalid bot. This is a bug.");
            StopMimicking(client);
        }
    }
    else
    {
        ReplyToCommand(client, "[Bot Mimic] You are not controlling any bot. Aim at a bot and press the mimic button to control it.");
    }
    
    return Plugin_Handled;
}

public void Event_PlayerSpawn(Event event, const char[] name, bool dontBroadcast)
{
    int client = GetClientOfUserId(event.GetInt("userid"));
    
    // If a player who was mimicking respawns, stop the mimicking
    if (!IsFakeClient(client) && g_bPlayerMimicking[client])
    {
        StopMimicking(client);
    }
}

public void Event_PlayerDeath(Event event, const char[] name, bool dontBroadcast)
{
    int client = GetClientOfUserId(event.GetInt("userid"));
    
    // If a player who was mimicking dies, stop the mimicking
    if (!IsFakeClient(client) && g_bPlayerMimicking[client])
    {
        StopMimicking(client);
    }
    
    // If a bot that was being mimicked dies, stop the mimicking for the controlling player
    int victim = GetClientOfUserId(event.GetInt("userid"));
    if (IsFakeClient(victim))
    {
        for (int i = 1; i <= MaxClients; i++)
        {
            if (IsClientInGame(i) && !IsFakeClient(i) && g_bPlayerMimicking[i] && g_iMimicBot[i] == victim)
            {
                StopMimicking(i);
            }
        }
    }
}

public void Event_RoundStart(Event event, const char[] name, bool dontBroadcast)
{
    // Reset all mimicking on round start
    for (int i = 1; i <= MaxClients; i++)
    {
        if (IsClientInGame(i) && !IsFakeClient(i) && g_bPlayerMimicking[i])
        {
            StopMimicking(i);
        }
    }
}

public void Event_RoundEnd(Event event, const char[] name, bool dontBroadcast)
{
    // Reset all mimicking on round end
    for (int i = 1; i <= MaxClients; i++)
    {
        if (IsClientInGame(i) && !IsFakeClient(i) && g_bPlayerMimicking[i])
        {
            StopMimicking(i);
        }
    }
}
