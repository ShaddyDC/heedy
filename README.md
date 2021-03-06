# Heedy

[![Docs](https://img.shields.io/badge/documentation-heedy.org-purple?style=flat-square)](https://heedy.org)[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/heedy/heedy?style=flat-square)](https://github.com/heedy/heedy/releases)[![PyPI](https://img.shields.io/pypi/v/heedy?style=flat-square)](https://pypi.org/project/heedy/)![GitHub Workflow Status](https://img.shields.io/github/workflow/status/heedy/heedy/Test?label=tests&style=flat-square)

**Note:** _Heedy is currently in alpha. You can try it out by downloading it from the releases page, but there is no guarantee that future versions will be backwards-compatible until full release._

A repository for your quantified-self data, and an extensible analysis engine.

There already exist many apps and fitness trackers that gather and attempt to make sense of your data. Most of these services are isolated - your phone's fitness tracking software knows nothing about your browser's time-tracking extension. Furthermore, each app and service has its own method for downloading data (if they offer raw data at all!), which makes an all-encompassing analysis of life extremely tedious. Heedy offers a self-hosted open-source way to put all of this data together into a single system.

Several existing aggregators already perform many of heedy's functions ([see the list here](https://github.com/woop/awesome-quantified-self#aggregators--dashboards)). However, they are all missing one of two critical components:

1. **Open-source and self-hosted**: Most existing tools are cloud-based, which means that all of your data is on "someone else's computer". While these companies may claim that they will not [sell your data](https://arstechnica.com/tech-policy/2019/03/ftc-investigates-whether-isps-sell-your-browsing-history-and-location-data/), or won't [turn it over to governments](<https://en.wikipedia.org/wiki/PRISM_(surveillance_program)>), they can change their minds (and terms of service) at any time. The only way to guarantee that your data will never be used against you is for it to be on your computer, operated by software you can audit yourself.
2. **Extensible**: Even a system with fantastic visualizations and powerful analysis has limited utility. This is because it can only show what the original authors assumed would be useful. Heedy offers a powerful plugin system - plugin writers can add new integrations, plots, or even modify core functionality with a few lines of python or javascript. A registry is planned, so that users can install plugins with the click of a button.

## Screenshots

The first screenshot is of sleep data uploaded by the [fitbit plugin](https://github.com/heedy/heedy-fitbit-plugin). The second is a jupyter notebook enabled by the [notebook plugin](https://github.com/heedy/heedy-notebook-plugin). Heedy's visualization and analysis capabilities are a work in progress, so there is a lot more to come!

[![Fitbit Plugin Example](./screenshot1.png)](https://github.com/heedy/heedy-fitbit-plugin)
[![Fitbit Notebook Example](./screenshot2.png)](https://github.com/heedy/heedy-notebook-plugin)

## Installing

1. Download the executable
2. Run the executable
3. Open your browser to http://localhost:1324

# Plugins

Heedy itself is very limited in scope. Most of its power comes from the plugins that allow you to integrate it with other services. Some plugins worth checking out:

- [fitbit](https://github.com/heedy/heedy-fitbit-plugin) - sync heedy with Fitbit, allowing you to access and analyze your wearables' data.
- [notebook](https://github.com/heedy/heedy-notebook-plugin) - analyze data directly within Heedy with Jupyter notebooks.

## Building

Building heedy requires at least go 1.15 and a recent version of node and npm.

### Release

```
git clone https://github.com/heedy/heedy
cd heedy
make
```

### Debug

```
git clone https://github.com/heedy/heedy
cd heedy
make debug
```

The debug version uses the assets from the `./assets` folder instead of embedding in the executable.

#### Watch frontend

To edit the frontend, you will want a debug build, and then run the following

```
make watch
```

This will watch all frontend files and rebuild them as they change, allowing you to edit them and see changes immediately by refreshing your browser.

### Verbose Mode

You can see everything heedy does, including all SQL statements and raw http requests by running it in verbose mode:

```
./heedy --verbose
```
