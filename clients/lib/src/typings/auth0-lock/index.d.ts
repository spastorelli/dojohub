declare module "auth0-lock" {

    interface Auth0Error {
        code: any;
        details: any;
        name: string;
        message: string;
        status: any;
    }

    interface Auth0LockPopupOptions {
        width: number;
        height: number;
        left: number;
        top: number;
    }

    interface Auth0LockOptions {
        authParams?: any;
        callbackURL?: string;
        connections?: string[];
        container?: string;
        closable?: boolean;
        dict?: any;
        defaultUserPasswordConnection?: string;
        defaultADUsernameFromEmailPrefix?: boolean;
        disableResetAction?: boolean;
        disableSignupAction?: boolean;
        focusInput?: boolean;
        forceJSONP?: boolean;
        gravatar?: boolean;
        integratedWindowsLogin?: boolean;
        icon?: string;
        loginAfterSignup?: boolean;
        popup?: boolean;
        popupOptions?: Auth0LockPopupOptions;
        rememberLastLogin?: boolean;
        resetLink?: string;
        responseType?: string;
        signupLink?: string;
        socialBigButtons?: boolean;
        sso?: boolean;
        theme?: string;
        usernameStyle?: any;
    }

    export interface Auth0LockStatic {
        new (clientId: string, domain: string, options: any): Auth0LockStatic;

        show(options: any): void;
        logout(callback: () => void): void;
        getProfile(token: string, callback: (error: Auth0Error, profile: Auth0UserProfile) => void) : void;

        on(event: "show", callback: () => void) : void;
        on(event: "hide", callback: () => void) : void;
        on(event: "unrecoverable_error", callback: (error: Auth0Error) => void) : void;
        on(event: "authorization_error", callback: (error: Auth0Error) => void) : void;
        on(event: "authenticated", callback: (authResult: any) => void) : void;
        on(event: string, callback: (...args: any[]) => void) : void;

        getClient(): Auth0Static;
    }

    let Auth0Lock: Auth0LockStatic;
    export default Auth0Lock;
}
