export default {
    sshReq: state => {
        let logintype = 0;
        if (state.sshInfo.s3KeyPath && state.sshInfo.s3KeyPath.trim()) {
            logintype = 2;
        } else if (state.sshInfo.privateKey && state.sshInfo.privateKey.trim()) {
            logintype = 1;
        }

        const sshInfo = {
            hostname: state.sshInfo.hostname,
            port: Number(state.sshInfo.port),
            username: state.sshInfo.username,
            logintype: logintype
        };
        if (state.sshInfo.password) {
            sshInfo.password = state.sshInfo.password;
        }
        if (logintype === 2) {
            sshInfo.s3KeyPath = state.sshInfo.s3KeyPath;
        } else if (logintype === 1) {
            sshInfo.privateKey = state.sshInfo.privateKey;
        }
        if (state.sshInfo.passphrase) {
            sshInfo.passphrase = state.sshInfo.passphrase;
        }
        const jsonStr = JSON.stringify(sshInfo);
        return window.btoa(jsonStr);
    }
}
