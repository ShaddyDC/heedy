package com.connectordb.dataconnect;


import android.util.Base64;
import android.util.Log;

import com.google.gson.JsonObject;

import retrofit.RequestInterceptor;
import retrofit.RestAdapter;
import retrofit.RetrofitError;
import retrofit.http.Body;
import retrofit.http.GET;
import retrofit.http.POST;
import retrofit.http.Path;
import retrofit.mime.TypedByteArray;
import retrofit.mime.TypedInput;

/**
 * Created by Daniel on 3/15/2015.
 */

public class ConnectorDB {
    private static final String TAG = "ConnectorDB";
    public final String user;
    public final String password;



    public class devicemaker {
        final String Name;
        devicemaker(String name) {
            this.Name=name;
        }
    }
    public class streammaker {
        final String Name;
        final String Type;
        streammaker(String name,String type) {
            this.Type=type;
            this.Name=name;
        }
    }

    public class genericresult {
        final String Status;
        final String Message;
        genericresult(String stat, String msg) {
            this.Status = stat;
            this.Message = msg;
        }
    }

    private interface DBAPI {
        @POST("/api/v1/json/{user}/device/")
        genericresult makedevice(@Path("user") String usr,@Body devicemaker mkr);

        @POST("/api/v1/json/byname/{user}/{device}/stream/")
        genericresult makestream(@Path("user") String user, @Path("device") String device, @Body streammaker mkr);

        @POST("/api/v1/json/byname/{user}/{device}/{stream}/point/")
        genericresult makestream(@Path("user") String user, @Path("device") String device,@Path("stream") String stream, @Body TypedInput obj);
    }

    private DBAPI dbapi;

    ConnectorDB(String url,String user, String password) {
        this.user = user;
        this.password = password;
        RequestInterceptor requestInterceptor = new RequestInterceptor() {
            @Override
            public void intercept(RequestInterceptor.RequestFacade request) {
                String credentials = ConnectorDB.this.user+":"+ConnectorDB.this.password;
                String basicauth = "Basic " + Base64.encodeToString(credentials.getBytes(), Base64.NO_WRAP);
                request.addHeader("Authorization", basicauth);
            }
        };
        RestAdapter restAdapter = new RestAdapter.Builder()
                .setEndpoint(url)
                .setRequestInterceptor(requestInterceptor)
                .build();
        this.dbapi = restAdapter.create(DBAPI.class);
    }
    void makedevice(String device) {
        try {
            this.dbapi.makedevice(this.user,new devicemaker(device));
        } catch (RetrofitError e) {
            Log.e(TAG,"makedevice:"+e.toString());
        }
    }
    void makestream(String device, String stream, String datatype) {
        try {
            this.dbapi.makestream(this.user,device,new streammaker(stream,datatype));
        } catch (RetrofitError e) {
            Log.e(TAG,"makestream:"+e.toString());
        }
    }

    boolean insert(String device, String stream, String jsondata) {
        try {
            TypedInput in = new TypedByteArray("application/json", jsondata.getBytes());
            this.dbapi.makestream(this.user,device,stream,in );
            return true;
        } catch (RetrofitError e) {
            Log.e(TAG,"insert:"+e.toString());
        }
        return false;
    }

}
