package com.connectordb.dataconnect;

import android.annotation.TargetApi;
import android.content.Context;
import android.content.Intent;
import android.content.res.Configuration;
import android.media.Ringtone;
import android.media.RingtoneManager;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.preference.ListPreference;
import android.preference.Preference;
import android.preference.PreferenceActivity;
import android.preference.PreferenceCategory;
import android.preference.PreferenceFragment;
import android.preference.PreferenceManager;
import android.preference.RingtonePreference;
import android.text.TextUtils;
import android.util.Log;


import java.util.List;

/**
 * A {@link PreferenceActivity} that presents a set of application settings. On
 * handset devices, settings are presented as a single list. On tablets,
 * settings are split by category, with category headers shown to the left of
 * the list of settings.
 * <p/>
 * See <a href="http://developer.android.com/design/patterns/settings.html">
 * Android Design: Settings</a> for design guidelines and the <a
 * href="http://developer.android.com/guide/topics/ui/settings.html">Settings
 * API Guide</a> for more information on developing a Settings UI.
 */
public class SettingsActivity extends PreferenceActivity {

    private static final String TAG = "SettingsActivity";


    /**
     * Determines whether to always show the simplified settings UI, where
     * settings are presented in a single list. When false, settings are shown
     * as a master/detail two-pane view on tablets. When true, a single pane is
     * shown on tablets.
     */
    private static final boolean ALWAYS_SIMPLE_PREFS = false;


    private Preference.OnPreferenceChangeListener sLocation = new Preference.OnPreferenceChangeListener() {
        @Override
        public boolean onPreferenceChange(Preference preference, Object value) {
            String stringValue = value.toString();
            Log.v(TAG, "Set GPS Logging: " + stringValue);

            // For list preferences, look up the correct display value in
            // the preference's 'entries' list.
            ListPreference listPreference = (ListPreference) preference;
            int index = listPreference.findIndexOfValue(stringValue);

            // Set the summary to reflect the new value.
            preference.setSummary(
                    index >= 0
                            ? listPreference.getEntries()[index]
                            : null);


            //Now notify the location service of the changed values
            Intent i = new Intent(SettingsActivity.this,LocationService.class);
            i.putExtra("location_update_frequency",Integer.parseInt(stringValue));
            startService(i);
            return true;
        }
    };

    private Preference.OnPreferenceChangeListener sMonitor = new Preference.OnPreferenceChangeListener() {
        @Override
        public boolean onPreferenceChange(Preference preference, Object value) {
            String stringValue = value.toString();
            Log.v(TAG, "Set Monitor Logging: " + stringValue);

            //Now set up the service
            Intent i = new Intent(SettingsActivity.this,MonitorService.class);
            i.putExtra("enabled",stringValue=="true");
            startService(i);
            return true;
        }
    };

    FitConnect fit;
    private Preference.OnPreferenceChangeListener sFit = new Preference.OnPreferenceChangeListener() {
        @Override
        public boolean onPreferenceChange(Preference preference, Object value) {
            String stringValue = value.toString();
            Log.v(TAG, "Set Fitness Logging: " + stringValue);

            if (stringValue=="true") {
                fit.subscribe();
            } else {
                fit.unsubscribe();
            }
            return true;
        }
    };


    private void setupLocation() {
        Preference location_pref = findPreference("location_frequency");
        String stringValue = PreferenceManager
                .getDefaultSharedPreferences(location_pref.getContext())
                .getString(location_pref.getKey(), "");

        sLocation.onPreferenceChange(location_pref,stringValue);

        location_pref.setOnPreferenceChangeListener(sLocation);

        Preference monitor_pref = findPreference("monitor_frequency");
        Boolean boolValue = PreferenceManager
                .getDefaultSharedPreferences(monitor_pref.getContext())
                .getBoolean(monitor_pref.getKey(), true);
        sMonitor.onPreferenceChange(monitor_pref,boolValue);
        monitor_pref.setOnPreferenceChangeListener(sMonitor);

        fit = new FitConnect(this,this,true);
        Preference fit_pref = findPreference("fitness_monitor");
        boolValue = PreferenceManager
                .getDefaultSharedPreferences(fit_pref.getContext())
                .getBoolean(fit_pref.getKey(), true);
        sFit.onPreferenceChange(monitor_pref,boolValue);
        fit_pref.setOnPreferenceChangeListener(sFit);

    }

    private void doStuff() {
        setupLocation();
        Preference button = (Preference)findPreference("syncbutton");
        button.setOnPreferenceClickListener(new Preference.OnPreferenceClickListener() {
            @Override
            public boolean onPreferenceClick(Preference arg0) {
                Preference db = findPreference("connectordb_server");
                String dbaddress = PreferenceManager
                        .getDefaultSharedPreferences(SettingsActivity.this)
                        .getString(db.getKey(), "");
                Preference usrname = findPreference("connectordb_username");
                String dbusr = PreferenceManager
                        .getDefaultSharedPreferences(SettingsActivity.this)
                        .getString(usrname.getKey(), "");
                Preference usrpass = findPreference("connectordb_password");
                String dbpass = PreferenceManager
                        .getDefaultSharedPreferences(SettingsActivity.this)
                        .getString(usrpass.getKey(), "");

                new asyncConnect(SettingsActivity.this).execute(dbaddress,dbusr,dbpass);
                return true;
            }
        });
        button = (Preference)findPreference("cachebutton");
        button.setOnPreferenceClickListener(new Preference.OnPreferenceClickListener() {
            @Override
            public boolean onPreferenceClick(Preference arg0) {
                DataCache c = DataCache.get(SettingsActivity.this);
                Preference button = (Preference)SettingsActivity.this.findPreference("cachebutton");
                button.setTitle(Integer.toString(c.Size()));
                return true;
            }
        });
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        if (requestCode == 1) {
            if (resultCode == RESULT_OK) {
                fit.reconnect();
            }
        }
    }

    @Override
    protected void onPostCreate(Bundle savedInstanceState) {
        Log.v(TAG, "PostCreate");
        super.onPostCreate(savedInstanceState);

        setupSimplePreferencesScreen();

        doStuff();
    }

    /**
     * Shows the simplified settings UI if the device configuration if the
     * device configuration dictates that a simplified, single-pane UI should be
     * shown.
     */
    private void setupSimplePreferencesScreen() {

        if (!isSimplePreferences(this)) {
            return;
        }

        // In the simplified UI, fragments are not used at all and we instead
        // use the older PreferenceActivity APIs.

        // Add 'general' preferences.
        addPreferencesFromResource(R.xml.pref_general);
        // Add GPS preferences
        PreferenceCategory fakeHeader = new PreferenceCategory(this);
        fakeHeader.setTitle(R.string.pref_header_sensors);
        getPreferenceScreen().addPreference(fakeHeader);
        addPreferencesFromResource(R.xml.pref_sensors);
        //Add sync prefs
        fakeHeader = new PreferenceCategory(this);
        fakeHeader.setTitle(R.string.pref_header_data_sync);
        getPreferenceScreen().addPreference(fakeHeader);
        addPreferencesFromResource(R.xml.pref_data_sync);
        // Add ConnectorDB preferences
        fakeHeader = new PreferenceCategory(this);
        fakeHeader.setTitle(R.string.pref_header_connectordb);
        getPreferenceScreen().addPreference(fakeHeader);
        addPreferencesFromResource(R.xml.pref_connectordb);


        // Bind the summaries of EditText/List/Dialog/Ringtone preferences to
        // their values. When their values change, their summaries are updated
        // to reflect the new value, per the Android Design guidelines.
        bindPreferenceSummaryToValue(findPreference("connectordb_username"));
        bindPreferenceSummaryToValue(findPreference("connectordb_server"));
        bindPreferenceSummaryToValue(findPreference("sync_wifi_frequency"));
        bindPreferenceSummaryToValue(findPreference("sync_mobile_frequency"));
    }

    /**
     * {@inheritDoc}
     */
    @Override
    public boolean onIsMultiPane() {
        return isXLargeTablet(this) && !isSimplePreferences(this);
    }

    /**
     * Helper method to determine if the device has an extra-large screen. For
     * example, 10" tablets are extra-large.
     */
    private static boolean isXLargeTablet(Context context) {
        return (context.getResources().getConfiguration().screenLayout
                & Configuration.SCREENLAYOUT_SIZE_MASK) >= Configuration.SCREENLAYOUT_SIZE_XLARGE;
    }

    /**
     * Determines whether the simplified settings UI should be shown. This is
     * true if this is forced via {@link #ALWAYS_SIMPLE_PREFS}, or the device
     * doesn't have newer APIs like {@link PreferenceFragment}, or the device
     * doesn't have an extra-large screen. In these cases, a single-pane
     * "simplified" settings UI should be shown.
     */
    private static boolean isSimplePreferences(Context context) {
        return ALWAYS_SIMPLE_PREFS
                || Build.VERSION.SDK_INT < Build.VERSION_CODES.HONEYCOMB
                || !isXLargeTablet(context);
    }

    /**
     * {@inheritDoc}
     */
    @Override
    @TargetApi(Build.VERSION_CODES.HONEYCOMB)
    public void onBuildHeaders(List<Header> target) {
        if (!isSimplePreferences(this)) {
            loadHeadersFromResource(R.xml.pref_headers, target);
        }
    }

    /**
     * A preference value change listener that updates the preference's summary
     * to reflect its new value.
     */
    private static Preference.OnPreferenceChangeListener sBindPreferenceSummaryToValueListener = new Preference.OnPreferenceChangeListener() {
        @Override
        public boolean onPreferenceChange(Preference preference, Object value) {
            String stringValue = value.toString();

            if (preference instanceof ListPreference) {
                // For list preferences, look up the correct display value in
                // the preference's 'entries' list.
                ListPreference listPreference = (ListPreference) preference;
                int index = listPreference.findIndexOfValue(stringValue);

                // Set the summary to reflect the new value.
                preference.setSummary(
                        index >= 0
                                ? listPreference.getEntries()[index]
                                : null);

            } else if (preference instanceof RingtonePreference) {
                // For ringtone preferences, look up the correct display value
                // using RingtoneManager.
                if (TextUtils.isEmpty(stringValue)) {
                    // Empty values correspond to 'silent' (no ringtone).
                    preference.setSummary(R.string.pref_ringtone_silent);

                } else {
                    Ringtone ringtone = RingtoneManager.getRingtone(
                            preference.getContext(), Uri.parse(stringValue));

                    if (ringtone == null) {
                        // Clear the summary if there was a lookup error.
                        preference.setSummary(null);
                    } else {
                        // Set the summary to reflect the new ringtone display
                        // name.
                        String name = ringtone.getTitle(preference.getContext());
                        preference.setSummary(name);
                    }
                }

            } else {
                // For all other preferences, set the summary to the value's
                // simple string representation.
                preference.setSummary(stringValue);
            }
            return true;
        }
    };

    /**
     * Binds a preference's summary to its value. More specifically, when the
     * preference's value is changed, its summary (line of text below the
     * preference title) is updated to reflect the value. The summary is also
     * immediately updated upon calling this method. The exact display format is
     * dependent on the type of preference.
     *
     * @see #sBindPreferenceSummaryToValueListener
     */
    private static void bindPreferenceSummaryToValue(Preference preference) {
        // Set the listener to watch for value changes.
        preference.setOnPreferenceChangeListener(sBindPreferenceSummaryToValueListener);

        // Trigger the listener immediately with the preference's
        // current value.
        sBindPreferenceSummaryToValueListener.onPreferenceChange(preference,
                PreferenceManager
                        .getDefaultSharedPreferences(preference.getContext())
                        .getString(preference.getKey(), ""));
    }

    /**
     * This fragment shows general preferences only. It is used when the
     * activity is showing a two-pane settings UI.
     */
    @TargetApi(Build.VERSION_CODES.HONEYCOMB)
    public static class GeneralPreferenceFragment extends PreferenceFragment {
        @Override
        public void onCreate(Bundle savedInstanceState) {
            super.onCreate(savedInstanceState);
            addPreferencesFromResource(R.xml.pref_general);

            // Bind the summaries of EditText/List/Dialog/Ringtone preferences
            // to their values. When their values change, their summaries are
            // updated to reflect the new value, per the Android Design
            // guidelines.
            //bindPreferenceSummaryToValue(findPreference("connectordb_username"));
            //bindPreferenceSummaryToValue(findPreference("connectordb_password"));
        }
    }
    @TargetApi(Build.VERSION_CODES.HONEYCOMB)
    public static class ConnectorDBPreferenceFragment extends PreferenceFragment {
        @Override
        public void onCreate(Bundle savedInstanceState) {
            super.onCreate(savedInstanceState);
            addPreferencesFromResource(R.xml.pref_connectordb);
            bindPreferenceSummaryToValue(findPreference("connectordb_server"));
            bindPreferenceSummaryToValue(findPreference("connectordb_username"));
        }
    }
    @TargetApi(Build.VERSION_CODES.HONEYCOMB)
    public static class SensorsPreferenceFragment extends PreferenceFragment {
        @Override
        public void onCreate(Bundle savedInstanceState) {
            super.onCreate(savedInstanceState);
            addPreferencesFromResource(R.xml.pref_sensors);
            //bindPreferenceSummaryToValue(findPreference("location_frequency"));
        }
    }
    /**
     * This fragment shows data and sync preferences only. It is used when the
     * activity is showing a two-pane settings UI.
     */
    @TargetApi(Build.VERSION_CODES.HONEYCOMB)
    public static class DataSyncPreferenceFragment extends PreferenceFragment {
        @Override
        public void onCreate(Bundle savedInstanceState) {
            super.onCreate(savedInstanceState);
            addPreferencesFromResource(R.xml.pref_data_sync);

            // Bind the summaries of EditText/List/Dialog/Ringtone preferences
            // to their values. When their values change, their summaries are
            // updated to reflect the new value, per the Android Design
            // guidelines.
            bindPreferenceSummaryToValue(findPreference("sync_wifi_frequency"));
            bindPreferenceSummaryToValue(findPreference("sync_mobile_frequency"));
        }
    }
}
