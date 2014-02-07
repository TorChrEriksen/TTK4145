mkdir $HOME/.vim/autoload
mkdir $HOME/.vim/autoload/go
mkdir $HOME/.vim/ftdetect
mkdir $HOME/.vim/ftplugin
mkdir $HOME/.vim/ftplugin/go
mkdir $HOME/.vim/indent
mkdir $HOME/.vim/plugin
ln -s /usr/local/go/misc/vim/autoload/go/complete.vim $HOME/.vim/autoload/go
ln -s /usr/local/go/misc/vim/ftdetect/gofiletype.vim $HOME/.vim/ftdetect
ln -s /usr/local/go/misc/vim/syntax/go.vim $HOME/.vim/syntax
ln -s /usr/local/go/misc/vim/ftplugin/go.vim $HOME/.vim/ftplugin
ln -s /usr/local/go/misc/vim/ftplugin/go/fmt.vim $HOME/.vim/ftplugin/go
ln -s /usr/local/go/misc/vim/ftplugin/go/import.vim $HOME/.vim/ftplugin/go
ln -s /usr/local/go/misc/vim/indent/go.vim $HOME/.vim/indent
touch $HOME/.vimrc
echo "set smartindent" >> $HOME/.vimrc
echo "set tabstop=4" >> $HOME/.vimrc
echo "set shiftwidth=4" >> $HOME/.vimrc
echo "set softtabstop=4" >> $HOME/.vimrc
echo "set expandtab" >> $HOME/.vimrc
