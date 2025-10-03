# Bot Name Changer - CSGO SourceMod Plugin

This SourceMod plugin automatically changes the names of bots in CSGO to a predefined list of professional player names.

## Features

- Automatically renames all bots in the server to professional player names
- Cycles through the name list for each new bot
- Works with existing bots and newly spawned bots
- Detects bots added via console commands
- Configurable to rename all bots or only default-named bots
- Force rename option to override any existing names
- **Hides the "Bot" or "电脑玩家" prefix** from bot names
- Supports multiple languages (English, Chinese, etc.)
- Multiple methods to hide bot prefixes for different server configurations

## Installation

1. Compile the plugin using the SourceMod compiler:
   ```
   spcomp bot_name_changer.sp
   ```

2. Copy the compiled `.smx` file to your server's plugins directory:
   ```
   /addons/sourcemod/plugins/
   ```

3. Copy the configuration file to your server's config directory:
   ```
   /addons/sourcemod/configs/
   ```

4. Restart your server or load the plugin using the following admin command:
   ```
   sm plugins load bot_name_changer
   ```

## Usage

The plugin automatically renames bots when:
- The map starts
- A bot is added to the server
- A bot spawns
- A round starts

You can also manually trigger bot renaming with the admin commands:
```
sm_renamebots                // Rename all bots
sm_forcerenamebot <target>   // Force rename a specific bot
sm_hidebotprefix             // Toggle hiding bot prefix
sm_changehidemethod <method> // Change the method used to hide bot prefix (0-4)
```

## Configuration

Edit the `bot_name_changer.cfg` file to configure the plugin:

```
// Enable debug mode (0 = disabled, 1 = enabled)
sm_botnamechanger_debug "1"

// Force rename bots even if they already have names (0 = disabled, 1 = enabled)
sm_botnamechanger_force "1"

// Rename all bots regardless of their current name (0 = disabled, 1 = enabled)
sm_botnamechanger_all "1"

// Interval in seconds for periodic bot renaming (1.0 - 60.0)
sm_botnamechanger_interval "10.0"

// Hide the 'Bot' prefix from bot names (0 = disabled, 1 = enabled)
sm_botnamechanger_hideprefix "1"

// Use invisible spaces to hide bot prefix (0 = disabled, 1 = enabled)
sm_botnamechanger_usespaces "1"

// Method to hide bot prefix (0-4)
// 0 = None (don't hide prefix)
// 1 = Spaces (use regular spaces)
// 2 = Invisible Characters (use zero-width spaces)
// 3 = Special Characters (use special Unicode characters)
// 4 = Mixed (combination of multiple methods)
sm_botnamechanger_hidemethod "4"
```

## Troubleshooting

If bots are not being renamed:

1. Verify the plugin is loaded:
   ```
   sm plugins list
   ```

2. Enable debug mode in the config file:
   ```
   sm_botnamechanger_debug "1"
   ```

3. Make sure these settings are enabled in the config:
   ```
   sm_botnamechanger_force "1"
   sm_botnamechanger_all "1"
   sm_botnamechanger_hideprefix "1"
   ```

4. Check the SourceMod logs for any error messages:
   ```
   logs/sourcemod/
   ```

5. Make sure you have admin privileges to use the manual rename command:
   ```
   sm_renamebots
   ```

6. Try reloading the plugin:
   ```
   sm plugins reload bot_name_changer
   ```

7. For local servers, make sure you have SourceMod properly installed and that the plugin is in the correct directory

8. If bots still aren't being renamed, try the force rename command on a specific bot:
   ```
   sm_forcerenamebot <bot_name>
   ```

9. If the bot prefix is still showing, try different hiding methods:
   ```
   sm_changehidemethod 1  // Try method 1 (spaces)
   sm_changehidemethod 2  // Try method 2 (invisible chars)
   sm_changehidemethod 3  // Try method 3 (special chars)
   sm_changehidemethod 4  // Try method 4 (mixed)
   ```

10. After changing the hide method, force rename all bots:
    ```
    sm_renamebots
    ```

## How Bot Prefix Hiding Works

The plugin uses several methods to hide the "Bot" prefix:

1. **Regular Spaces**: Adds multiple spaces at the beginning of the name.

2. **Invisible Unicode Characters**: Uses zero-width spaces and other invisible Unicode characters.

3. **Special Characters**: Uses specific Unicode control characters that might interfere with the prefix system.

4. **Mixed Method**: Combines multiple approaches for maximum effectiveness.

The plugin continuously monitors bot names and fixes them if the game adds the prefix back.

## Bot Names List

The plugin includes the following professional player names:

- benjyfishy
- Bugha
- EpikWhale
- Clix
- Aqua
- Nyhrox
- MrSavage
- Mongraal
- Arkhram
- Zayt
- Saf
- ZexRow
- Khanada
- Reverse2k
- Deyy
- Mero
- JannisZ
- Tayson
- Vadeal
- Kami
- 3xPO
- Brax1n
- xii
- ReaL
- yomamx
- Jazz
- wipeer
- LiiLii
- zYK
- Revenge
- Davi
- xvx
- Mystik
- Rembrandt
- Sya
- Axiyo
- Lei
- Dvl
- M1ka
- Mendoza

## Customization

If you want to use different names, you can modify the `g_szProPlayerNames` array in the `bot_name_changer.sp` file and recompile the plugin.

If you encounter a bot prefix in a language not currently supported, you can add it to the `g_szBotPrefixes` array in the source code. 