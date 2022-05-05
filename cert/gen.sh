#!/bin/bash

# 1. 生成CA私钥和自签名证书,用于签署CSR
rm *.pem
# 使用RSA4096生成私钥和证书请求(-x509 选项输出证书而不是请求),指定有效天数为365天,并将创建的私钥写入ca-key.pem文件中,并将证书写入ca-cert.pem文件中 -nodes 用于测试时不需要输入密码
openssl req -x509 -newkey rsa:4096 -days 365 -keyout ca-key.pem -out ca-cert.pem -nodes -subj "/C=CN/ST=Shaanxi/L=Xian/O=sta/OU=xupt/CN=sta/emailAddress=1647193241@qq.com"
echo "CA私钥和自签名证书生成完毕"

# 查看证书信息-noout 不输出原始编码值，-text以可读的方式查看
openssl x509 -in ca-cert.pem -noout -text

# 2.1 生成一个用于服务器的私钥和CSR证书请求
openssl req -newkey rsa:4096 -keyout server-key.pem -out server-req.pem -nodes -subj "/C=CN/ST=Shaanxi/L=Xian/O=server/OU=server/CN=server/emailAddress=1647193241@qq.com"
echo "服务器私钥和CSR证书请求生成完毕"

# 2.2 使用CA的私钥去注册服务器CSR，以及返回注册的证书
# 使用 -req 表示传入一个CSR证书请求，-in 输入请求文件，-CA 传递CA证书文件，-CAkey 传递CA私钥 -CAcreateserial 自动生成唯一的证书序号 -out 输出的证书文件名 -days 有效天数 -extfile 额外配置文件 http://man.openbsd.org/x509v3.cnf
openssl x509 -req -in server-req.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -days 365 -extfile server-ext.cnf
echo "服务器证书生成完毕"
# 查看证书信息-noout 不输出原始编码值，-text以可读的方式查看
openssl x509 -in server-cert.pem -noout -text

# 2.3 验证证书是否有效
# -CAfile 传入信任的CA证书 和 需要验证的服务器证书
openssl verify -CAfile ca-cert.pem server-cert.pem

# 3.1 生成一个用于客户端的私钥和CSR证书请求
openssl req -newkey rsa:4096 -keyout client-key.pem -out client-req.pem -nodes -subj "/C=CN/ST=Shaanxi/L=Xian/O=client/OU=client/CN=client/emailAddress=1647193241@qq.com"
echo "客户端私钥和CSR证书请求生成完毕"

# 3.2 使用CA的私钥去注册服务器CSR，以及返回注册的证书
# 使用 -req 表示传入一个CSR证书请求，-in 输入请求文件，-CA 传递CA证书文件，-CAkey 传递CA私钥 -CAcreateserial 自动生成唯一的证书序号 -out 输出的证书文件名 -days 有效天数 -extfile 额外配置文件 http://man.openbsd.org/x509v3.cnf
openssl x509 -req -in client-req.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -days 365 -extfile client-ext.cnf
echo "客户端证书生成完毕"
# 查看证书信息-noout 不输出原始编码值，-text以可读的方式查看
openssl x509 -in client-cert.pem -noout -text

# 3.3 验证证书是否有效
# -CAfile 传入信任的CA证书 和 需要验证的服务器证书
openssl verify -CAfile ca-cert.pem client-cert.pem
