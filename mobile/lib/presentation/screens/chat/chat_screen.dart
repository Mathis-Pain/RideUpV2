import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:supabase_flutter/supabase_flutter.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

import 'package:rideup/core/config/env.dart';

class ChatScreen extends StatefulWidget {
  const ChatScreen({super.key, required this.eventId, required this.eventTitle});

  final String eventId;
  final String eventTitle;

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final _msgCtrl = TextEditingController();
  final _scrollCtrl = ScrollController();
  final _supabase = Supabase.instance.client;

  List<Map<String, dynamic>> _messages = [];
  bool _loading = true;
  bool _sending = false;

  WebSocketChannel? _ws;
  StreamSubscription? _wsSub;

  String get _userId => _supabase.auth.currentUser!.id;
  String get _username =>
      _supabase.auth.currentUser?.userMetadata?['username'] as String? ?? '?';

  @override
  void initState() {
    super.initState();
    _loadMessages();
    _connectWs();
  }

  @override
  void dispose() {
    _msgCtrl.dispose();
    _scrollCtrl.dispose();
    _wsSub?.cancel();
    _ws?.sink.close();
    super.dispose();
  }

  Future<void> _loadMessages() async {
    final data = await _supabase
        .from('messages')
        .select('id, content, created_at, user_id, profiles(username)')
        .eq('event_id', widget.eventId)
        .order('created_at')
        .limit(50);

    if (mounted) {
      setState(() {
        _messages = List<Map<String, dynamic>>.from(data);
        _loading = false;
      });
      _scrollToBottom();
    }
  }

  void _connectWs() {
    final token = _supabase.auth.currentSession?.accessToken;
    if (token == null) return;

    final wsBase = Env.apiBaseUrl.replaceFirst('http', 'ws');
    final uri = Uri.parse('$wsBase/ws/chat/${widget.eventId}?token=$token');

    _ws = WebSocketChannel.connect(uri);
    _wsSub = _ws!.stream.listen(
      (raw) {
        final data = jsonDecode(raw as String) as Map<String, dynamic>;
        if (mounted) {
          setState(() => _messages.add({
                'user_id': data['sender_id'],
                'content': data['body'],
                'created_at': data['sent_at'],
                'profiles': {'username': data['username']},
              }));
          _scrollToBottom();
        }
      },
      onError: (e) => debugPrint('ws error: $e'),
      onDone: () => debugPrint('ws closed'),
    );
  }

  Future<void> _send() async {
    final text = _msgCtrl.text.trim();
    if (text.isEmpty || _ws == null) return;

    setState(() => _sending = true);
    _msgCtrl.clear();

    try {
      _ws!.sink.add(jsonEncode({'body': text, 'username': _username}));
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Erreur : $e')));
      }
    } finally {
      if (mounted) setState(() => _sending = false);
    }
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollCtrl.hasClients) {
        _scrollCtrl.animateTo(
          _scrollCtrl.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(widget.eventTitle)),
      body: Column(
        children: [
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : _messages.isEmpty
                    ? const Center(child: Text('Aucun message — soyez le premier !'))
                    : ListView.builder(
                        controller: _scrollCtrl,
                        padding: const EdgeInsets.symmetric(
                            horizontal: 12, vertical: 8),
                        itemCount: _messages.length,
                        itemBuilder: (_, i) => _MessageBubble(
                          msg: _messages[i],
                          isMe: _messages[i]['user_id'] == _userId,
                        ),
                      ),
          ),
          _InputBar(
            controller: _msgCtrl,
            sending: _sending,
            onSend: _send,
          ),
        ],
      ),
    );
  }
}

class _MessageBubble extends StatelessWidget {
  const _MessageBubble({required this.msg, required this.isMe});

  final Map<String, dynamic> msg;
  final bool isMe;

  @override
  Widget build(BuildContext context) {
    final profile = msg['profiles'] as Map<String, dynamic>?;
    final username = profile?['username'] as String? ?? '?';
    final content = msg['content'] as String? ?? '';
    final colors = Theme.of(context).colorScheme;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment:
            isMe ? MainAxisAlignment.end : MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          if (!isMe) ...[
            CircleAvatar(
              radius: 14,
              child: Text(username[0].toUpperCase(),
                  style: const TextStyle(fontSize: 12)),
            ),
            const SizedBox(width: 6),
          ],
          Flexible(
            child: Column(
              crossAxisAlignment:
                  isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start,
              children: [
                if (!isMe)
                  Padding(
                    padding: const EdgeInsets.only(left: 4, bottom: 2),
                    child: Text(username,
                        style: Theme.of(context).textTheme.labelSmall),
                  ),
                Container(
                  padding: const EdgeInsets.symmetric(
                      horizontal: 12, vertical: 8),
                  decoration: BoxDecoration(
                    color: isMe ? colors.primary : colors.surfaceContainerHigh,
                    borderRadius: BorderRadius.only(
                      topLeft: const Radius.circular(16),
                      topRight: const Radius.circular(16),
                      bottomLeft: Radius.circular(isMe ? 16 : 4),
                      bottomRight: Radius.circular(isMe ? 4 : 16),
                    ),
                  ),
                  child: Text(
                    content,
                    style: TextStyle(
                      color: isMe ? colors.onPrimary : colors.onSurface,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _InputBar extends StatelessWidget {
  const _InputBar({
    required this.controller,
    required this.sending,
    required this.onSend,
  });

  final TextEditingController controller;
  final bool sending;
  final VoidCallback onSend;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(12, 8, 12, 8),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: controller,
                textInputAction: TextInputAction.send,
                onSubmitted: (_) => onSend(),
                decoration: InputDecoration(
                  hintText: 'Message…',
                  filled: true,
                  contentPadding: const EdgeInsets.symmetric(
                      horizontal: 16, vertical: 10),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(24),
                    borderSide: BorderSide.none,
                  ),
                ),
              ),
            ),
            const SizedBox(width: 8),
            IconButton.filled(
              onPressed: sending ? null : onSend,
              icon: sending
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.send),
            ),
          ],
        ),
      ),
    );
  }
}
