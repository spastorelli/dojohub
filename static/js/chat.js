/**
 * The chat application.
 *
 * @param {string} host The websocket server hostname.
 * @param {string} appClientId The application client id.
 * @param {string} chatElemId The ID of element that holds the
 *      DOM for chat application.
 *
 * @constructor
 */
var ChatApp = function(host, appClientId, chatElemId) {
    var chatDom = $('#' + chatElemId);
    this.chatConvElem_ = chatDom.find('#chat-conv')[0];
    this.chatForm_ = chatDom.find('#chat-form');
    this.chatMsg_ = this.chatForm_.find('#chat-msg');
    this.appClient_ = new coderdojo.dojohub.AppClient(host, appClientId);
};


/**
 * Starts the Chat application.
 */
ChatApp.prototype.start = function() {
      this.chatMsg_.keypress(this.onEnterKeyPress_.bind(this));
      this.chatForm_.submit(this.onFormSubmit_.bind(this));
      this.appClient_.onClose = this.onWsClose_.bind(this);
      this.appClient_.onMessage = this.onWsMessage_.bind(this);
      this.appClient_.connect();
};


/**
 * Submits the chat message when enter key is pressed.
 * @param {event} keyEvent The key event.
 * @return {boolean}
 * @private
 */
ChatApp.prototype.onEnterKeyPress_ = function(keyEvent) {
    if (keyEvent.which === 13) {
        this.chatForm_.submit();
        return false;
    }
};


/**
 * Handles the form submission.
 * @return {boolean}
 * @private
 */
ChatApp.prototype.onFormSubmit_ = function() {
    var msgData = this.chatMsg_.val();
    if (this.appClient_.isClosed() || !msgData) {
        return false;
    }
    this.appClient_.send(msgData);
    this.updateChat_(msgData, /** opt_incoming */ false);
    this.chatMsg_.val('');
    return false;
};


/**
 * Handles when the websocket is closed.
 * @param {event} wsEvent The websocket onclose event.
 * @private
 */
ChatApp.prototype.onWsClose_ = function(wsEvent) {
    var chatMsgElem_ = this.chatMsg_[0];
    chatMsgElem_.readOnly = true;
    this.updateChat_('Connection closed');
};


/**
 * Handles websocket incoming messages.
 * @param {event} wsEvent The websocket onmessage event.
 * @private
 */
ChatApp.prototype.onWsMessage_ = function(wsEvent) {
    var msgData = wsEvent.data;
    this.updateChat_(msgData, /** opt_incoming */ true);
};


/**
 * Updates the chat conversation thread with the provided message.
 *
 * @param {string} message The message to add to the conversation thread.
 * @param {boolean} opt_incoming Whether the message is incoming or outcoming.
 * @private
 */
ChatApp.prototype.updateChat_ = function(message, opt_incoming) {
    var incoming = (opt_incoming === undefined) ? true : opt_incoming;
    var msgClass = incoming ? 'chat-msg-in' : 'chat-msg-out';
    var msgContainer = $('<span class="mdl-chip ' + msgClass + '">');
    var msgChip = $('<span class="mdl-chip__text">');
    msgChip.text(message);
    msgChip.appendTo(msgContainer);
    msgContainer.appendTo(this.chatConvElem_);
    this.chatConvElem_.scrollTop = this.chatConvElem_.scrollHeight;
};
