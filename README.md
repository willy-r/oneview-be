# OneView

OneView é um sistema de mensagens efêmeras onde cada mensagem só pode ser visualizada **uma única vez** e depois desaparece para sempre, inspirado em conceitos de privacidade.

## 🚀 Tecnologias
- **Go** com Fiber (backend REST e WebSocket)
- **JWT** para autenticação
- **GORM + SQLite** para persistência
- **AES** para criptografia simétrica de mensagens
- **WebSocket** para notificações em tempo real

## 📦 Funcionalidades
- Cadastro e login de usuários
- Envio de mensagens com criptografia
- Visualização única de mensagens
- Notificação em tempo real ao destinatário e ao remetente via WebSocket

## 🔐 Segurança
- Mensagens criptografadas com AES-128
- Autenticação com JWT
- Comunicação WebSocket entre cliente e servidor com identificação via `userID`

## 🧪 Testes
Você pode testar as APIs com ferramentas como Insomnia ou Postman, e WebSocket com o Insomnia ou no app em React Native.

## 📂 Estrutura de diretórios
```
oneview/
├── cmd/            # main.go (entrada do servidor)
├── internal/
│   ├── handler/    # lógica de handlers HTTP e WS
│   ├── model/      # modelos do GORM
│   ├── middleware/ # middleware de autenticação
├── pkg/config/     # inicialização do banco
```

## ▶️ Como rodar localmente
```bash
go mod tidy
go run ./cmd/main.go
```

## 📄 Licença
Este projeto está sob a licença MIT. Veja o conteúdo em `LICENSE`.
