# CoinWatch
A tool to keep track of your tokens across different protocols. Please copy config.sample.yml to ./config.yml to start, then edit based on your needs. You can change location of the DB 
which defaults to $HOME/.coinwatch.db and configuration path using parameters, check ```coinwatch --help``` for a full 
list of options. 

Inside the configuration every wallet gets data from a balance provider, currently the app supports the following
options:
- `subscan` provides data on most substrate based tokens
- `kraken` supports balance from Kraken exchange (any token)
- `algoexplorer` currently support balance for algo token only
- `minaexplorer` mina token balance
- `blockcypher` bitcoin balance

### Telegram bot
The tool is meant to be run as a Telegram bot, it will provide a nice visualization of your tokens, start the bot using
```bash
coinwatch -v bot --chat-id YOURCHATID --token YOURTELEGRAMTOKEN 
```
Right now supported commands are /summary <days> and /allocation

Summary will output something like
```
Update
30 May 22 15:36 +0200
Balance
Token Price   EUR    1D       1M     
GLMR  1.21€   989€   -2.7%    -0.1%    
AZERO 0.948€  295€   -1.4%    +0.3%    
ASTR  0.059€  78€    +2.4%    +1%    
DOT   9.47€   317€   -3.4%    +5% 
MOVR  20.8€   254€   +16%     +0.5%    
Summary
Total 1933€          -2.6%  +25%        
Performance
 1990 ┤ ╭─╮
 1980 ┼╮│ │   ╭╮
 1970 ┤││ ╰───╯╰╮
 1960 ┤╰╯       ╰╮  ╭╮
 1950 ┤          ╰╮╭╯╰─╮╭─╮
 1940 ┤           ╰╯   ╰╯ ╰╮
 1930 ┤                    │ ╭─
 1920 ┤                    ╰╮│
 1910 ┤                     ╰╯
```

While allocation will show the actual token allocation
```
Update
30 May 22 14:15 UTC
Allocation
T     Pct  Bal    Price  1D    1W    
GLMR  70%  818    1.23€  +7.9% +4.2% 
AZERO 20%  312    0.917€ +2.5% -8.8% 
ASTR  5.6% 1.2K   0.061€ +13%  +0%   
DOT   2.2% 335    9.59€  +4.0% -1.5% 
KSM   0.4% 8.54   70.3€  +4.7% -11%  
```

### Docker
A ready made Docker image is available at Docker hub, just do:
```bash
docker pull johnuopini/coinwatch:latest
```
