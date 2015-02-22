package com.connectordb.dataconnect;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.util.Log;

import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.GooglePlayServicesUtil;

public class SetupReceiver extends BroadcastReceiver {

    private static final String TAG = "SetupReceiver";

    public SetupReceiver() {
    }

    @Override
    public void onReceive(Context context, Intent intent) {
        // TODO: This method is called when the BroadcastReceiver is receiving
        // an Intent broadcast.
        /*On boot, start the necessary services:
        Intent locationIntent = new Intent(context, LocationService.class);
        context.startService(locationIntent);
        */

        //throw new UnsupportedOperationException("Not yet implemented");
        Log.v(TAG,"RECEIVED BOOT");

        //Start the GPS location service
        Intent i = new Intent(context,LocationService.class);
        i.putExtra("value","0");
        context.startService(i);
    }

    /**
     * Check the device to make sure it has the Google Play Services APK. If
     * it doesn't, display a dialog that allows users to download the APK from
     * the Google Play Store or enable it in the device's system settings.
     */
    /*
    private final static int PLAY_SERVICES_RESOLUTION_REQUEST = 9000;
    private boolean checkPlayServices() {
        int resultCode = GooglePlayServicesUtil.isGooglePlayServicesAvailable(this);
        if (resultCode != ConnectionResult.SUCCESS) {
            if (GooglePlayServicesUtil.isUserRecoverableError(resultCode)) {
                GooglePlayServicesUtil.getErrorDialog(resultCode, this,
                        PLAY_SERVICES_RESOLUTION_REQUEST).show();
            } else {
                Log.i(TAG, "This device is not supported.");
                finish();
            }
            return false;
        }
        return true;
    }*/

    /*Should probably register for GCM here...*/
}
