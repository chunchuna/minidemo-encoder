/**
 * Bot Name Changer - CSGO SourceMod Plugin
 * Changes the names of bots to a predefined list of professional player names
 * Hides the "Bot" or "电脑玩家" prefix from bot names
 */

#include <sourcemod>
#include <sdktools>
#include <cstrike>

#pragma semicolon 1
#pragma newdecls required

public Plugin myinfo = {
    name = "Bot Name Changer",
    author = "Admin",
    description = "Changes bot names to professional player names and hides bot prefix",
    version = "1.0.3",
    url = ""
};

// ConVars
ConVar g_cvDebugMode;
ConVar g_cvForceRename;
ConVar g_cvRenameAllBots;
ConVar g_cvRenameInterval;
ConVar g_cvHidePrefix;
ConVar g_cvUseSpaces;
ConVar g_cvHideMethod;

// Array of professional player names
char g_szProPlayerNames[][] = {
    "benjyfishy", "Bugha", "EpikWhale", "Clix", "Aqua", "Nyhrox", "MrSavage", 
    "Mongraal", "Arkhram", "Zayt", "Saf", "ZexRow", "Khanada", "Reverse2k", 
    "Deyy", "Mero", "JannisZ", "Tayson", "Vadeal", "Kami", "3xPO", "Brax1n", 
    "xii", "ReaL", "yomamx", "Jazz", "wipeer", "LiiLii", "zYK", "Revenge", 
    "Davi", "xvx", "Mystik", "Rembrandt", "Sya", "Axiyo", "Lei", "Dvl", 
    "M1ka", "Mendoza"
};

// Known bot prefixes in different languages
char g_szBotPrefixes[][] = {
    "BOT ", "Bot ", "电脑玩家 ", "機器人 ", "Bote ", "Ботик ", "ボット ", "봇 "
};

// Special characters for prefix hiding
char g_szSpecialChars[][] = {
    "⁣", "​", "‌", "‍", "‎", "‏", "⁠", "⁡", "⁢", "⁣", "⁤", "⁦", "⁧", "⁨", "⁩", "⁪", "⁫", "⁬", "⁭",
    "\u200B", "\u200C", "\u200D", "\u200E", "\u200F", "\u2060", "\u2061", "\u2062", "\u2063", "\u2064",
    "\u206A", "\u206B", "\u206C", "\u206D", "\u206E", "\u206F"
};

// Current index for cycling through names
int g_iCurrentNameIndex = 0;

// Store bot names to avoid duplicates
char g_szBotNames[MAXPLAYERS + 1][MAX_NAME_LENGTH];

// Handle for name update timer
Handle g_hNameUpdateTimer = null;

// 在全局变量区域添加这些变量
bool g_bNameChanged[MAXPLAYERS + 1]; // 跟踪bot名字是否已经修改

// Plugin initialization
public void OnPluginStart()
{
    // Create ConVars
    g_cvDebugMode = CreateConVar("sm_botnamechanger_debug", "1", "Enable debug mode for Bot Name Changer plugin", _, true, 0.0, true, 1.0);
    g_cvForceRename = CreateConVar("sm_botnamechanger_force", "1", "Force rename bots even if they already have names", _, true, 0.0, true, 1.0);
    g_cvRenameAllBots = CreateConVar("sm_botnamechanger_all", "1", "Rename all bots regardless of their current name", _, true, 0.0, true, 1.0);
    g_cvRenameInterval = CreateConVar("sm_botnamechanger_interval", "10.0", "Interval in seconds for periodic bot renaming", _, true, 1.0, true, 60.0);
    g_cvHidePrefix = CreateConVar("sm_botnamechanger_hideprefix", "1", "Hide the 'Bot' prefix from bot names", _, true, 0.0, true, 1.0);
    g_cvUseSpaces = CreateConVar("sm_botnamechanger_usespaces", "1", "Use invisible spaces to hide bot prefix", _, true, 0.0, true, 1.0);
    g_cvHideMethod = CreateConVar("sm_botnamechanger_hidemethod", "2", "Method to hide bot prefix (0=none, 1=spaces, 2=invisible chars, 3=special chars, 4=mixed)", _, true, 0.0, true, 4.0);
    
    // Register commands
    RegAdminCmd("sm_renamebots", Command_RenameBots, ADMFLAG_GENERIC, "Manually trigger bot renaming");
    RegAdminCmd("sm_forcerenamebot", Command_ForceRenameBot, ADMFLAG_GENERIC, "Force rename a specific bot");
    RegAdminCmd("sm_hidebotprefix", Command_HideBotPrefix, ADMFLAG_GENERIC, "Toggle hiding bot prefix");
    RegAdminCmd("sm_changehidemethod", Command_ChangeHideMethod, ADMFLAG_GENERIC, "Change the method used to hide bot prefix");
    
    // Hook events
    HookEvent("player_connect_full", Event_PlayerConnectFull, EventHookMode_Post);
    HookEvent("player_spawn", Event_PlayerSpawn, EventHookMode_Post);
    HookEvent("round_start", Event_RoundStart, EventHookMode_Post);
    HookEvent("player_team", Event_PlayerTeam, EventHookMode_Post);
    HookEvent("player_death", Event_PlayerDeath, EventHookMode_Post);
    
    // Hook bot commands to catch bot creation
    AddCommandListener(Command_Bot, "bot_add");
    AddCommandListener(Command_Bot, "bot_add_t");
    AddCommandListener(Command_Bot, "bot_add_ct");
    AddCommandListener(Command_Bot, "bot_place");
    
    // Create timer to check for bots periodically (backup method)
    CreateTimer(g_cvRenameInterval.FloatValue, Timer_CheckBots, _, TIMER_REPEAT);
    
    // Create name update timer (runs more frequently)
    g_hNameUpdateTimer = CreateTimer(1.0, Timer_UpdateNames, _, TIMER_REPEAT);
    
    // Log plugin initialization
    PrintToServer("[Bot Name Changer] Plugin initialized");
    LogMessage("Bot Name Changer plugin initialized");
    
    // Auto-execute config
    AutoExecConfig(true, "bot_name_changer");
    
    // Check for late load
    if (IsServerProcessing())
    {
        CreateTimer(3.0, Timer_LateLoad);
    }
    
    // 初始化名字修改跟踪
    for (int i = 1; i <= MaxClients; i++)
    {
        g_bNameChanged[i] = false;
    }
}

public void OnPluginEnd()
{
    // Clean up timers
    if (g_hNameUpdateTimer != null)
    {
        KillTimer(g_hNameUpdateTimer);
        g_hNameUpdateTimer = null;
    }
}

// Late load timer
public Action Timer_LateLoad(Handle timer)
{
    DebugPrint("Late load detected, renaming existing bots");
    RenameBots(true);
    return Plugin_Stop;
}

// Called when the map starts
public void OnMapStart()
{
    // Reset the name index on map start
    g_iCurrentNameIndex = 0;
    
    // Clear the bot names array
    for (int i = 1; i <= MaxClients; i++)
    {
        g_szBotNames[i][0] = '\0';
    }
    
    // Create timer to rename bots at map start (after a short delay)
    CreateTimer(5.0, Timer_RenameBotsOnMapStart);
    
    DebugPrint("Map started, will rename bots shortly");
}

// Called when a client connects
public void OnClientConnected(int client)
{
    if (IsFakeClient(client))
    {
        DebugPrint("Bot connected: %d - Will attempt to rename shortly", client);
        CreateTimer(1.0, Timer_RenameBot, client);
    }
}

// Called when a client puts in server
public void OnClientPutInServer(int client)
{
    if (IsFakeClient(client))
    {
        DebugPrint("Bot put in server: %d - Will attempt to rename shortly", client);
        CreateTimer(1.0, Timer_RenameBot, client);
    }
    else
    {
        // 如果是bot，重置名字修改状态
        g_bNameChanged[client] = false;
    }
}

// Command handler for manual bot renaming
public Action Command_RenameBots(int client, int args)
{
    bool forceRename = true;
    RenameBots(forceRename);
    ReplyToCommand(client, "[Bot Name Changer] Manually triggered bot renaming (forced: %s)", forceRename ? "yes" : "no");
    return Plugin_Handled;
}

// Command handler for forcing rename on a specific bot
public Action Command_ForceRenameBot(int client, int args)
{
    if (args < 1)
    {
        ReplyToCommand(client, "Usage: sm_forcerenamebot <target>");
        return Plugin_Handled;
    }
    
    char arg[64];
    GetCmdArg(1, arg, sizeof(arg));
    
    int target = FindTarget(client, arg, true, false);
    if (target == -1)
    {
        return Plugin_Handled;
    }
    
    if (!IsFakeClient(target))
    {
        ReplyToCommand(client, "[Bot Name Changer] Target is not a bot!");
        return Plugin_Handled;
    }
    
    RenameBot(target, true);
    ReplyToCommand(client, "[Bot Name Changer] Forced rename on bot %d", target);
    
    return Plugin_Handled;
}

// Command handler for toggling bot prefix hiding
public Action Command_HideBotPrefix(int client, int args)
{
    bool newValue = !g_cvHidePrefix.BoolValue;
    g_cvHidePrefix.SetBool(newValue);
    
    ReplyToCommand(client, "[Bot Name Changer] Bot prefix hiding %s", newValue ? "enabled" : "disabled");
    
    // Trigger a rename with the new setting
    RenameBots(true);
    
    return Plugin_Handled;
}

// Command handler for changing the hide method
public Action Command_ChangeHideMethod(int client, int args)
{
    if (args < 1)
    {
        ReplyToCommand(client, "Usage: sm_changehidemethod <method>");
        ReplyToCommand(client, "0: None, 1: Spaces, 2: Invisible Chars, 3: Special Chars, 4: Mixed");
        return Plugin_Handled;
    }

    char arg[8];
    GetCmdArg(1, arg, sizeof(arg));
    int newMethod = StringToInt(arg);
    
    if (newMethod < 0 || newMethod > 4)
    {
        ReplyToCommand(client, "Invalid method number. Must be 0-4.");
        return Plugin_Handled;
    }

    g_cvHideMethod.SetInt(newMethod);
    ReplyToCommand(client, "[Bot Name Changer] Hide method changed to: %d", newMethod);
    RenameBots(true); // Apply the new method immediately

    return Plugin_Handled;
}

// Timer callback to rename bots at map start
public Action Timer_RenameBotsOnMapStart(Handle timer)
{
    DebugPrint("Map start timer triggered, renaming bots");
    RenameBots(true);
    return Plugin_Stop;
}

// Timer callback to periodically check for bots
public Action Timer_CheckBots(Handle timer)
{
    DebugPrint("Periodic check for bots");
    RenameBots(g_cvForceRename.BoolValue);
    return Plugin_Continue;
}

// Timer callback to update bot names (runs more frequently)
public Action Timer_UpdateNames(Handle timer)
{
    // Only check for prefix issues
    for (int i = 1; i <= MaxClients; i++)
    {
        if (IsValidClient(i) && IsFakeClient(i) && g_szBotNames[i][0] != '\0')
        {
            char currentName[MAX_NAME_LENGTH];
            GetClientName(i, currentName, sizeof(currentName));
            
            // Check if the name has a bot prefix
            bool hasPrefix = false;
            for (int j = 0; j < sizeof(g_szBotPrefixes); j++)
            {
                if (StrContains(currentName, g_szBotPrefixes[j]) == 0)
                {
                    hasPrefix = true;
                    break;
                }
            }
            
            // If it has a prefix but shouldn't, fix it
            if (hasPrefix && g_cvHidePrefix.BoolValue)
            {
                FixBotName(i);
            }
        }
    }
    
    return Plugin_Continue;
}

// Event when a player fully connects
public Action Event_PlayerConnectFull(Event event, const char[] name, bool dontBroadcast)
{
    int client = GetClientOfUserId(event.GetInt("userid"));
    
    // Check if the client is valid and is a bot
    if (IsValidClient(client) && IsFakeClient(client))
    {
        DebugPrint("Bot connect_full event for client %d", client);
        // Delay the rename slightly to ensure the bot is fully connected
        CreateTimer(0.5, Timer_RenameBot, client);
    }
    
    return Plugin_Continue;
}

// Event when a player spawns
public Action Event_PlayerSpawn(Event event, const char[] name, bool dontBroadcast)
{
    int client = GetClientOfUserId(event.GetInt("userid"));
    
    // Check if the client is valid and is a bot
    if (IsValidClient(client) && IsFakeClient(client))
    {
        DebugPrint("Bot spawn event for client %d", client);
        // Delay the rename slightly
        CreateTimer(0.2, Timer_RenameBot, client);
    }
    
    return Plugin_Continue;
}

// Event when a round starts
public Action Event_RoundStart(Event event, const char[] name, bool dontBroadcast)
{
    DebugPrint("Round start event, checking bots");
    CreateTimer(1.0, Timer_CheckBotsAfterRoundStart);
    return Plugin_Continue;
}

// Event when a player changes team
public Action Event_PlayerTeam(Event event, const char[] name, bool dontBroadcast)
{
    int client = GetClientOfUserId(event.GetInt("userid"));
    
    // Check if the client is valid and is a bot
    if (IsValidClient(client) && IsFakeClient(client))
    {
        DebugPrint("Bot team change event for client %d", client);
        // Delay the rename slightly
        CreateTimer(0.2, Timer_RenameBot, client);
    }
    
    return Plugin_Continue;
}

// Event when a player dies
public Action Event_PlayerDeath(Event event, const char[] name, bool dontBroadcast)
{
    int client = GetClientOfUserId(event.GetInt("userid"));
    
    // Check if the client is valid and is a bot
    if (IsValidClient(client) && IsFakeClient(client))
    {
        DebugPrint("Bot death event for client %d", client);
        // Check if name needs fixing after respawn
        CreateTimer(0.5, Timer_CheckBotNameAfterDeath, client);
    }
    
    return Plugin_Continue;
}

// Timer callback to check bot name after death
public Action Timer_CheckBotNameAfterDeath(Handle timer, any client)
{
    if (IsValidClient(client) && IsFakeClient(client) && g_szBotNames[client][0] != '\0')
    {
        char currentName[MAX_NAME_LENGTH];
        GetClientName(client, currentName, sizeof(currentName));
        
        // Check if the name has changed or has a bot prefix
        if (!StrEqual(currentName, g_szBotNames[client]))
        {
            FixBotName(client);
        }
    }
    
    return Plugin_Stop;
}

// Timer callback to check bots after round start
public Action Timer_CheckBotsAfterRoundStart(Handle timer)
{
    RenameBots(g_cvForceRename.BoolValue);
    return Plugin_Stop;
}

// Timer callback to rename a specific bot
public Action Timer_RenameBot(Handle timer, any client)
{
    if (IsValidClient(client) && IsFakeClient(client))
    {
        char currentName[MAX_NAME_LENGTH];
        GetClientName(client, currentName, sizeof(currentName));
        
        DebugPrint("Checking bot %d with current name '%s'", client, currentName);
        
        // Check if this bot should be renamed
        bool shouldRename = ShouldRenameBot(client, currentName);
        
        if (shouldRename)
        {
            RenameBot(client, g_cvForceRename.BoolValue);
        }
        else
        {
            DebugPrint("Bot %d with name '%s' does not need renaming", client, currentName);
            
            // Still check if we need to hide the prefix
            if (g_cvHidePrefix.BoolValue)
            {
                // Check if the name has a bot prefix
                bool hasPrefix = false;
                for (int j = 0; j < sizeof(g_szBotPrefixes); j++)
                {
                    if (StrContains(currentName, g_szBotPrefixes[j]) == 0)
                    {
                        hasPrefix = true;
                        break;
                    }
                }
                
                if (hasPrefix)
                {
                    FixBotName(client);
                }
            }
        }
    }
    
    return Plugin_Stop;
}

// Command listener for bot creation commands
public Action Command_Bot(int client, const char[] command, int args)
{
    // Bot command was used, schedule a check for new bots
    DebugPrint("Bot command detected: %s", command);
    CreateTimer(1.0, Timer_CheckBotsAfterCommand);
    return Plugin_Continue;
}

// Timer callback to check for bots after a bot command
public Action Timer_CheckBotsAfterCommand(Handle timer)
{
    DebugPrint("Checking for bots after bot command");
    RenameBots(g_cvForceRename.BoolValue);
    return Plugin_Stop;
}

// Function to check if a bot should be renamed
bool ShouldRenameBot(int client, const char[] currentName)
{
    // If rename all bots is enabled, always return true
    if (g_cvRenameAllBots.BoolValue)
    {
        return true;
    }
    
    // Otherwise, only rename bots with default names
    return (StrContains(currentName, "BOT", false) != -1 || 
            StrContains(currentName, "bot", false) != -1 ||
            StrEqual(currentName, "Counter-Terrorist", false) ||
            StrEqual(currentName, "Terrorist", false));
}

// Function to rename all bots in the server
void RenameBots(bool forceRename = false)
{
    DebugPrint("Starting to rename all bots (force: %s)", forceRename ? "yes" : "no");
    
    int botsFound = 0;
    int botsRenamed = 0;
    
    for (int i = 1; i <= MaxClients; i++)
    {
        if (IsValidClient(i) && IsFakeClient(i))
        {
            botsFound++;
            char currentName[MAX_NAME_LENGTH];
            GetClientName(i, currentName, sizeof(currentName));
            
            DebugPrint("Found bot %d with name '%s'", i, currentName);
            
            // Check if this bot should be renamed
            bool shouldRename = ShouldRenameBot(i, currentName);
            
            if (shouldRename || forceRename)
            {
                RenameBot(i, forceRename);
                botsRenamed++;
            }
            else
            {
                DebugPrint("Bot %d with name '%s' does not need renaming", i, currentName);
                
                // Still check if we need to hide the prefix
                if (g_cvHidePrefix.BoolValue)
                {
                    // Check if the name has a bot prefix
                    bool hasPrefix = false;
                    for (int j = 0; j < sizeof(g_szBotPrefixes); j++)
                    {
                        if (StrContains(currentName, g_szBotPrefixes[j]) == 0)
                        {
                            hasPrefix = true;
                            break;
                        }
                    }
                    
                    if (hasPrefix)
                    {
                        FixBotName(i);
                        botsRenamed++;
                    }
                }
            }
        }
    }
    
    DebugPrint("Rename all bots complete. Found: %d, Renamed: %d", botsFound, botsRenamed);
}

// Function to rename a specific bot
void RenameBot(int client, bool forceRename = false)
{
    if (!IsValidClient(client) || !IsFakeClient(client))
        return;
    
    char oldName[MAX_NAME_LENGTH];
    GetClientName(client, oldName, sizeof(oldName));
    
    // Check if we should rename this bot
    if (!forceRename && !ShouldRenameBot(client, oldName))
    {
        DebugPrint("Skipping rename for bot %d with name '%s' (not forced)", client, oldName);
        return;
    }
    
    // Get the next name from the list
    char newName[MAX_NAME_LENGTH];
    GetNextBotName(newName, sizeof(newName));
    
    // If hiding prefix is enabled, add invisible characters
    char finalName[MAX_NAME_LENGTH];
    if (g_cvHidePrefix.BoolValue)
    {
        ApplyPrefixHiding(newName, finalName, sizeof(finalName));
    }
    else
    {
        strcopy(finalName, sizeof(finalName), newName);
    }
    
    // Set the bot's name
    SetClientName(client, finalName);
    
    // Store the new name (without prefix)
    strcopy(g_szBotNames[client], MAX_NAME_LENGTH, newName);
    
    // Force name change to take effect immediately
    CS_SwitchTeam(client, GetClientTeam(client));
    CS_UpdateClientModel(client);
    
    // Double check if name was changed
    char verifyName[MAX_NAME_LENGTH];
    GetClientName(client, verifyName, sizeof(verifyName));
    
    // Log the name change
    DebugPrint("Renamed bot %d from '%s' to '%s' (verify: '%s')", client, oldName, newName, verifyName);
    LogMessage("Renamed bot %d from '%s' to '%s'", client, oldName, newName);
    
    // Print to server console for visibility
    PrintToServer("[Bot Name Changer] Renamed bot from '%s' to '%s'", oldName, newName);
    
    // Print to all players
    PrintToChatAll("[Bot Name Changer] Bot '%s' is now known as '%s'", oldName, newName);
    
    // Schedule a check to make sure the name sticks
    CreateTimer(0.5, Timer_VerifyBotName, client);
}

// Function to fix a bot's name (hide prefix)
void FixBotName(int client)
{
    if (!IsValidClient(client) || !IsFakeClient(client) || g_szBotNames[client][0] == '\0')
        return;
    
    char currentName[MAX_NAME_LENGTH];
    GetClientName(client, currentName, sizeof(currentName));
    
    // Check if the name already matches what we want
    if (StrEqual(currentName, g_szBotNames[client]))
        return;
    
    // Check if the name has a bot prefix
    bool hasPrefix = false;
    for (int j = 0; j < sizeof(g_szBotPrefixes); j++)
    {
        if (StrContains(currentName, g_szBotPrefixes[j]) == 0)
        {
            hasPrefix = true;
            break;
        }
    }
    
    if (hasPrefix || !StrEqual(currentName, g_szBotNames[client]))
    {
        char finalName[MAX_NAME_LENGTH];
        
        if (g_cvHidePrefix.BoolValue)
        {
            ApplyPrefixHiding(g_szBotNames[client], finalName, sizeof(finalName));
        }
        else
        {
            strcopy(finalName, sizeof(finalName), g_szBotNames[client]);
        }
        
        // Set the bot's name
        SetClientName(client, finalName);
        
        // Force name change to take effect immediately
        CS_SwitchTeam(client, GetClientTeam(client));
        CS_UpdateClientModel(client);
        
        DebugPrint("Fixed bot %d name from '%s' to '%s'", client, currentName, g_szBotNames[client]);
    }
}

// Timer to verify bot name after renaming
public Action Timer_VerifyBotName(Handle timer, any client)
{
    if (IsValidClient(client) && IsFakeClient(client) && g_szBotNames[client][0] != '\0')
    {
        char currentName[MAX_NAME_LENGTH];
        GetClientName(client, currentName, sizeof(currentName));
        
        // Check if the name has a bot prefix
        bool hasPrefix = false;
        for (int j = 0; j < sizeof(g_szBotPrefixes); j++)
        {
            if (StrContains(currentName, g_szBotPrefixes[j]) == 0)
            {
                hasPrefix = true;
                break;
            }
        }
        
        // If it has a prefix but shouldn't, fix it
        if (hasPrefix && g_cvHidePrefix.BoolValue)
        {
            FixBotName(client);
        }
        // If the name doesn't match what we want
        else if (!StrEqual(currentName, g_szBotNames[client]) && !hasPrefix)
        {
            FixBotName(client);
        }
    }
    
    return Plugin_Stop;
}

// Function to apply prefix hiding to a name
void ApplyPrefixHiding(const char[] originalName, char[] buffer, int bufferSize)
{
    int hideMethod = g_cvHideMethod.IntValue;

    switch (hideMethod)
    {
        case 0: // None
            strcopy(buffer, bufferSize, originalName);
            break;
        case 1: // Spaces
            Format(buffer, bufferSize, "     %s", originalName);
            break;
        case 2: // Invisible Chars
            Format(buffer, bufferSize, "⁣⁣⁣⁣⁣​‌‍‎‏⁠⁡⁢⁣⁤⁦⁧⁨⁩⁪⁫⁬⁭%s", originalName);
            break;
        case 3: // Special Chars
        {
            // Use a mix of special characters that might interfere with the prefix system
            char specialPrefix[64];
            
            // Create a special prefix using various special characters
            Format(specialPrefix, sizeof(specialPrefix), "%c%c%c%c%c", 
                0x200B, 0x200C, 0x200D, 0x200E, 0x200F);
                
            // Add some more special characters
            StrCat(specialPrefix, sizeof(specialPrefix), "\u2060\u2061\u2062\u2063");
            
            // Format the final name
            Format(buffer, bufferSize, "%s%s", specialPrefix, originalName);
            break;
        }
        case 4: // Mixed (Method that combines multiple approaches)
        {
            // Start with some spaces
            strcopy(buffer, bufferSize, "  ");
            
            // Add some invisible characters
            StrCat(buffer, bufferSize, "⁣⁣⁣​‌‍");
            
            // Add some Unicode control characters
            StrCat(buffer, bufferSize, "\u200B\u200C\u200D\u200E\u200F");
            
            // Add the original name
            StrCat(buffer, bufferSize, originalName);
            break;
        }
        default: // Fallback to spaces if method is invalid
            Format(buffer, bufferSize, "     %s", originalName);
            break;
    }
}

// Function to get the next bot name from the list
void GetNextBotName(char[] name, int maxLength)
{
    // Get the current name
    strcopy(name, maxLength, g_szProPlayerNames[g_iCurrentNameIndex]);
    
    // Increment index and wrap around if needed
    g_iCurrentNameIndex = (g_iCurrentNameIndex + 1) % sizeof(g_szProPlayerNames);
    
    DebugPrint("Selected bot name: %s (index: %d)", name, g_iCurrentNameIndex);
}

// Helper function to check if a client is valid
bool IsValidClient(int client)
{
    return (client > 0 && client <= MaxClients && IsClientConnected(client) && IsClientInGame(client));
}

// Helper function for debug prints
void DebugPrint(const char[] format, any ...)
{
    if (g_cvDebugMode.BoolValue)
    {
        char buffer[512];
        VFormat(buffer, sizeof(buffer), format, 2);
        PrintToServer("[Bot Name Changer DEBUG] %s", buffer);
        LogMessage("[DEBUG] %s", buffer);
    }
} 

// 添加一个函数来设置bot的团队颜色
void SetBotTeamColor(int bot)
{
    // 获取当前游戏模式
    ConVar mp_teamlogo_1 = FindConVar("mp_teamlogo_1");
    ConVar mp_teamlogo_2 = FindConVar("mp_teamlogo_2");
    
    if (mp_teamlogo_1 != null && mp_teamlogo_2 != null)
    {
        // 随机选择一个颜色标识
        char colorTag[16];
        int colorChoice = GetRandomInt(0, 4);
        
        switch (colorChoice)
        {
            case 0: strcopy(colorTag, sizeof(colorTag), "red");
            case 1: strcopy(colorTag, sizeof(colorTag), "blue");
            case 2: strcopy(colorTag, sizeof(colorTag), "yellow");
            case 3: strcopy(colorTag, sizeof(colorTag), "green");
            case 4: strcopy(colorTag, sizeof(colorTag), "orange");
        }
        
        // 设置团队标志，这会影响头像和小地图颜色
        int team = GetClientTeam(bot);
        if (team == 2) // T队
        {
            mp_teamlogo_1.SetString(colorTag);
        }
        else if (team == 3) // CT队
        {
            mp_teamlogo_2.SetString(colorTag);
        }
        
        // 设置玩家颜色标签
        CS_SetClientClanTag(bot, colorTag);
    }
} 