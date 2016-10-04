"use strict"

import Auth0Lock from "auth0-lock";

/**
 * The DojoHub Auth0 application URL.
 */
const Auth0Url: string  = "dojohub.eu.auth0.com";


/**
 * The DojoHub login icon.
 */
const Auth0LoginIcon: string =
    "//upload.wikimedia.org/wikipedia/commons/9/99/CoderDojo_Logo.png";

export module dojohub {
    /**
     * Defines a Dojohub application client.
     */
    export class AppClient {

        /**
         * The DojoHub Auth0 application ID.
         */
        private auth0AppId: string;

        /**
         * The websocket connection instance.
         */
        private wsConn: WebSocket;

        /**
         * The DojoHub websocket server URL.
         */
        private wsUrl: string;

        /**
         * Event handlers called on WebSocket events to be implemented by consumers.
         */
        public onOpen:(ev:Event) => void = function (event:Event) {};
        public onClose:(ev:CloseEvent) => void = function (event:CloseEvent) {};
        public onMessage:(ev:MessageEvent) => void = function (event:MessageEvent) {};
        public onError:(ev:ErrorEvent) => void = function (event:ErrorEvent) {};

        /**
         * @param dojohubUrl  The address of the DojoHub server.
         * @param appId       The DojoHub Auth0 application ID.
         * @constructor
         */
        constructor(dojohubUrl: string, appId: string) {
            this.auth0AppId = appId;
            this.wsUrl = "ws://" + dojohubUrl + "/ws/";
        };

        /**
         * Initializes the WebSocket connection to the Message Hub.
         */
        private initWsConnection(token: string) {
            this.wsUrl += "?t=" + token;
            this.wsConn = new WebSocket(this.wsUrl);

            this.wsConn.onopen = (event: Event) => {
                console.log("Connected to the Message Hub.");
                this.onOpen(event);
            };
            this.wsConn.onclose = (event: CloseEvent) => {
                console.log("Connection to the Message Hub closed.");
                this.onClose(event);
                this.wsConn = null;
            };
            this.wsConn.onmessage = (event: MessageEvent) => {
                console.log("Message received from the Message Hub.");
                this.onMessage(event);
            };
            this.wsConn.onerror = (event: ErrorEvent) => {
                console.log("Error received from the Message Hub." + event)
            };
        };

        /**
         * Connects the application client to the DojoHub WebSocket server.
         */
        public connect() {
            var options = {auth: {redirect: false}};
            var lock = new Auth0Lock(this.auth0AppId, Auth0Url, options);
            lock.on("authenticated", (authResult: any) => {
              // c TODO(spastorelli): Store token to localStorage.
              this.initWsConnection(authResult.idToken);
            });
            lock.on("authorization_error", (error: Auth0Error) => {
              throw "Error occured during authentication: " + error.message;
            });
            lock.show({icon: Auth0LoginIcon});
        };

        /**
         * Sends a message to the Message Hub.
         * TODO(spastorelli): Defines a data type for the messages.
         */
        public send(data: any) {
            console.log("Sending message to the Message Hub.")
            this.wsConn.send(data);
        };

        /**
         * Checks if the connection is closed.
         */
        public isClosed() {
            return this.wsConn === null;
        };
    }
}
