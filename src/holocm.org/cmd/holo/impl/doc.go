/*******************************************************************************
*
* Copyright 2017 Luke Shumaker <lukeshu@parabola.nu>
*
* This file is part of Holo.
*
* Holo is free software: you can redistribute it and/or modify it under the
* terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* Holo is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* Holo. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

// Package impl implements most of the "holo" program; as a set of
// modular pieces.
//
// Each piece is usable without the others; there is no global or
// shared state between them... mostly.  The "output" package does
// provide some share global state; but it all takes place in the
// "output" package, not here.
//
// 1. Call "NewConfigReader(rootDir)" to get a LineReader that will
// read lines from all of the relevant configuration files.
//
// 2. Call "ReadConfig(LineReader)" to parse holorc configuration.
//
// 3. Call "NewRuntimeManager(rootDir)" to set up the runtime cache
// directory.
//
// 4. Call "runtimeManager.GetPlugins(config.Plugins,
// YOUR_PLUGIN_LOADER)" to load all of the configued plugins.
// Alternatively, you may loop over config.Plugins yourself, calling
// "NewPluginHandle()" for each plugin, with a runtime you got from
// "runtimeManager.NewRuntime(pluginID)".
//
// 5. Generate a list of entities you want to operate on, represented
// by a "[]*EntityHandle".  You probably want to do this by first
// getting a list of all entities from your plugins, then filtering
// that list to the entities you want.
//
// You can get a list of all entities from your plugins with
// "GetAllEntities(pluginHandles)"; alternatively, you can loop over
// your plugin handles yourself, calling ".Scan()" on each.
//
// Then, you can filter that list with "FilterEntities(entities,
// selectors)"; alternatively, you can loop over your entities and
// selectors calling entityHandle.MatchesSelector(selector) for each.
//
// 6. Call one of "CommandApply", "CommandDiff", or "CommandScan" with
// your list of entities.
//
// Beware that you should probably have a system-wide mutex/lockfile,
// ensuring that only one CommandApply is running on a system at once!
// You can do this with a PID file at a fixed location with
// AcquirePidFile(), and calling pidFile.Release() when done.
//
// Alternatively, you can loop over the list of entities yourself,
// calling ".Apply()", ".PrintReport()", ".PrintScanReport()", or
// ".RenderDiff()" on each.
//
// 7. Call "runtimeManager.Close()" to clean up.
//
//     ----
//
// All that remains for the calling program to do is option parsing
// and user-interface concerns.
package impl
