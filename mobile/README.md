# Hub phone app

An Expo (React Native) app for Android and iOS. It shows the Hub web interface's phone layout full screen, exactly as the web
draws it, with the same theme and every feature except Calls management. The phone handles what a browser would:

- **Sign-in bound to the phone.** The app makes an Ed25519 key on first run. The private half stays in the phone's secure
  storage (Keychain or Android Keystore), marked "this device only", and it is never backed up. The web page never sees it.
  - Sign-in sends the public key with a signature.
  - Every token refresh must be signed by that key.
  - A refresh token copied to another device fails, and using it ends the sign-in (`device_unverified`).
- **System notifications** for what the page announces over its live connection while the app is in the background.
- **Files** (downloads, exports, attachments) open in the phone's open/save/share sheet.
- **Links that leave Hub** open in the phone's browser.
- **Other phone behaviour:** the Android back button, the camera and microphone for calls, meetings and voice notes, the
  clipboard, and status-bar colours that follow the light and dark themes.

All data comes from the Hub backend through the web interface; the app stores nothing of its own.

## Set up

```sh
cd mobile
npm install
cp .env.example .env    # then set EXPO_PUBLIC_HUB_WEB_URL (the site) and EXPO_PUBLIC_HUB_API_URL (the backend), both https
```

The address must be the **https** site people open in a browser. End-to-end encryption needs a secure page, so a plain
`http://192.168…` dev server will not work on a phone. Deploy the frontend first, because the app shows whatever that site
serves.

## Run on a phone (Expo Go)

```sh
npx expo start
```

Scan the QR code with the camera (iOS) or the Expo Go app (Android). Use the latest Expo Go (SDK 58).

## Push notifications

New messages, incoming calls and notifications reach the phone while Hub is in the background or closed: in the notification
area and on the lock screen, with the Hub icon, like a chat app. A tap opens the chat or the notifications. They do not work
in Expo Go on Android; use a build (below). Chat messages are end-to-end encrypted, so a push never shows what was written;
the organization's "details in emails" setting decides whether it names the sender and the chat.

Android (Firebase Cloud Messaging):

1. In the [Firebase console](https://console.firebase.google.com), create a project and add an Android app with the package
   `com.newgen.hub`. Download `google-services.json` into `mobile/` (it is not committed). For EAS builds, upload it as a
   file secret instead: `npx eas-cli@latest env:create --name GOOGLE_SERVICES_JSON --type file --value ./google-services.json`.
2. In Project settings > Service accounts, generate a private key and set its JSON as `HUB_FCM_CREDENTIALS` on the backend.

iOS (Apple Push Notification service): create an APNs key in the Apple developer account and set `HUB_APNS_KEY`,
`HUB_APNS_KEY_ID` and `HUB_APNS_TEAM_ID` on the backend (`HUB_APNS_SANDBOX=true` with `HUB_APS_ENV=development` for
development builds).

## Build an APK

With EAS (cloud build; needs a free Expo account):

```sh
npx eas-cli@latest login
npm run apk            # Android APK to install directly
npm run ios:build      # iOS build (needs an Apple developer account)
```

Locally (needs the Android SDK and JDK 17+):

```sh
npm run apk:local      # -> android/app/build/outputs/apk/release/app-release.apk
```

For the Play Store, set up your own signing key; a local build signs with the debug key. Change `com.newgen.hub` in
`app.json` to your own identifier before publishing.
