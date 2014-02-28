FILE=$HOME/.vimrc

echo "Creating directories..."
mkdir -p --verbose $HOME/.vim/autoload/go
mkdir -p --verbose $HOME/.vim/ftdetect
mkdir -p --verbose $HOME/.vim/ftplugin/go
mkdir -p --verbose $HOME/.vim/indent
mkdir -p --verbose $HOME/.vim/syntax

if [ -f "$FILE" ]; then
    echo "File $FILE exists, config is appended to file:"
else
    touch $FILE
    echo "Writing config to $FILE:"
fi

echo "syntax on" | tee -a $FILE
echo "filetype plugin on" | tee -a $FILE
echo "filetype indent on" | tee -a $FILE
echo "set smartindent" | tee -a $FILE
echo "set tabstop=4" | tee -a $FILE
echo "set shiftwidth=4" | tee -a $FILE
echo "set softtabstop=4" | tee -a $FILE
echo "set expandtab" | tee -a $FILE

echo "Creating symbolic links..."
ln -s /usr/local/go/misc/vim/autoload/go/complete.vim $HOME/.vim/autoload/go
ln -s /usr/local/go/misc/vim/ftdetect/gofiletype.vim $HOME/.vim/ftdetect
ln -s /usr/local/go/misc/vim/syntax/go.vim $HOME/.vim/syntax
ln -s /usr/local/go/misc/vim/ftplugin/go.vim $HOME/.vim/ftplugin
ln -s /usr/local/go/misc/vim/ftplugin/go/fmt.vim $HOME/.vim/ftplugin/go
ln -s /usr/local/go/misc/vim/ftplugin/go/import.vim $HOME/.vim/ftplugin/go
ln -s /usr/local/go/misc/vim/indent/go.vim $HOME/.vim/indent

echo "vim Go setup complete!"
