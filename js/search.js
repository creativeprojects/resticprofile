window.relearn = window.relearn || {};

window.relearn.runInitialSearch = function(){
    if( window.relearn.isSearchInit && window.relearn.isLunrInit ){
        searchDetail();
    }
}

var lunrIndex, pagesIndex;

function initLunrIndex( index ){
    pagesIndex = index;
    // Set up lunrjs by declaring the fields we use
    // Also provide their boost level for the ranking
    lunrIndex = lunr(function() {
        this.use(lunr.multiLanguage.apply(null, contentLangs));
        this.ref('index');
        this.field('title', {
            boost: 15
        });
        this.field('tags', {
            boost: 10
        });
        this.field('content', {
            boost: 5
        });

        this.pipeline.remove(lunr.stemmer);
        this.searchPipeline.remove(lunr.stemmer);

        // Feed lunr with each file and let lunr actually index them
        pagesIndex.forEach(function(page, idx) {
            page.index = idx;
            this.add(page);
        }, this);
    });

    window.relearn.isLunrInit = true;
    window.relearn.runInitialSearch();
}

function triggerSearch(){
    searchDetail();
    var input = document.querySelector('#search-by-detail');
    if( !input ){
        return;
    }
    var value = input.value;
    var url = new URL( window.location );
    var oldValue = url.searchParams.get('search-by');
    if( value != oldValue ){
        url.searchParams.set('search-by', value);
        window.history.pushState(url.toString(), '', url);
    }
}

window.addEventListener('popstate', function ( event ) {
    // restart search if browsed thru history
    if (event.state && event.state.indexOf('search.html?search-by=') >= 0) {
        window.location.reload();
    }
});

var input = document.querySelector('#search-by-detail');
if( input ){
    input.addEventListener( 'keydown', function(event) {
        // if we are pressing ESC in the searchdetail our focus will
        // be stolen by the other event handlers, so we have to refocus
        // here after a short while
        if (event.key == "Escape") {
            setTimeout( function(){ input.focus(); }, 0 );
        }
    });
}

function initLunrJson() {
    // old way to load the search index via XHR;
    // this does not work if pages are served via
    // file:// protocol; this is only left for
    // backward compatiblity if the user did not
    // define the SEARCH output format for the homepage
    if( window.index_json_url && !window.index_js_url ){
        $.getJSON(index_json_url)
        .done(function(index) {
            initLunrIndex(index);
        })
        .fail(function(jqxhr, textStatus, error) {
            var err = textStatus + ', ' + error;
            console.error('Error getting Hugo index file:', err);
        });
    }
}

function initLunrJs() {
    // new way to load our search index
    if( window.index_js_url ){
        var js = document.createElement("script");
        js.src = index_js_url;
        js.setAttribute("async", "");
        js.onload = function(){
            initLunrIndex(relearn_search_index);
        };
        js.onerror = function(e){
            console.error('Error getting Hugo index file');
        };
        document.head.appendChild(js);
    }
}

/**
 * Trigger a search in lunr and transform the result
 *
 * @param  {String} term
 * @return {Array}  results
 */
function search(term) {
    // Find the item in our index corresponding to the lunr one to have more info
    // Remove Lunr special search characters: https://lunrjs.com/guides/searching.html
    var searchTerm = lunr.tokenizer(term.replace(/[*:^~+-]/, ' ')).reduce( function(a,token){return a.concat(searchPatterns(token.str))}, []).join(' ');
    return !searchTerm || !lunrIndex ? [] : lunrIndex.search(searchTerm).map(function(result) {
        return { index: result.ref, matches: Object.keys(result.matchData.metadata) }
    });
}

function searchPatterns(word) {
    // for short words high amounts of typos doesn't make sense
    // for long words we allow less typos because this largly increases search time
    var typos = [
        { len:  -1, typos: 1 },
        { len:  60, typos: 2 },
        { len:  40, typos: 3 },
        { len:  20, typos: 4 },
        { len:  16, typos: 3 },
        { len:  12, typos: 2 },
        { len:   8, typos: 1 },
        { len:   4, typos: 0 },
    ];
    return [
        word + '^100',
        word + '*^10',
        '*' + word + '^10',
        word + '~' + typos.reduce( function( a, c, i ){ return word.length < c.len ? c : a; } ).typos + '^1'
    ];
}


function resolvePlaceholders( s, args ) {
    var args = args || [];
    // use replace to iterate over the string
    // select the match and check if the related argument is present
    // if yes, replace the match with the argument
    return s.replace(/{([0-9]+)}/g, function (match, index) {
        // check if the argument is present
        return typeof args[index] == 'undefined' ? match : args[index];
    });
};

function searchDetail() {
    var input = document.querySelector('#search-by-detail');
    if( !input ){
        return;
    }
    var value = input.value;
    var results = document.querySelector('#searchresults');
    var hint = document.querySelector('.searchhint');
    hint.innerText = '';
    results.textContent = '';
    var a = search( value );
    if( a.length ){
        hint.innerText = resolvePlaceholders( window.T_N_results_found, [ value, a.length ] );
        a.forEach( function(item){
            var page = pagesIndex[item.index];
            var numContextWords = 10;
            var contextPattern = '(?:\\S+ +){0,' + numContextWords + '}\\S*\\b(?:' +
                item.matches.map( function(match){return match.replace(/\W/g, '\\$&')} ).join('|') +
                ')\\b\\S*(?: +\\S+){0,' + numContextWords + '}';
            var context = page.content.match(new RegExp(contextPattern, 'i'));
            var divcontext = document.createElement('div');
            divcontext.className = 'context';
            divcontext.innerText = (context || '');
            var divsuggestion = document.createElement('a');
            divsuggestion.className = 'autocomplete-suggestion';
            divsuggestion.setAttribute('data-term', value);
            divsuggestion.setAttribute('data-title', page.title);
            divsuggestion.setAttribute('href', baseUri + page.uri);
            divsuggestion.setAttribute('data-context', context);
            divsuggestion.innerText = '» ' + page.title;
            divsuggestion.appendChild(divcontext);
            results.appendChild( divsuggestion );
        });
        window.relearn.markSearch();
    }
    else if( value.length ) {
        hint.innerText = resolvePlaceholders( window.T_No_results_found, [ value ] );
    }
    input.focus();
    setTimeout( adjustContentWidth, 0 );
}

// Let's get started
initLunrJson();
initLunrJs();
$(function() {
    var url = new URL( window.location );
    window.history.replaceState(url.toString(), '', url);

    var searchList = new autoComplete({
        /* selector for the search box element */
        selectorToInsert: '#header-wrapper',
        selector: '#search-by',
        /* source is the callback to perform the search */
        source: function(term, response) {
            response(search(term));
        },
        /* renderItem displays individual search results */
        renderItem: function(item, term) {
            var page = pagesIndex[item.index];
            var numContextWords = 2;
            var contextPattern = '(?:\\S+ +){0,' + numContextWords + '}\\S*\\b(?:' +
                item.matches.map( function(match){return match.replace(/\W/g, '\\$&')} ).join('|') +
                ')\\b\\S*(?: +\\S+){0,' + numContextWords + '}';
            var context = page.content.match(new RegExp(contextPattern, 'i'));
            var divcontext = document.createElement('div');
            divcontext.className = 'context';
            divcontext.innerText = (context || '');
            var divsuggestion = document.createElement('div');
            divsuggestion.className = 'autocomplete-suggestion';
            divsuggestion.setAttribute('data-term', term);
            divsuggestion.setAttribute('data-title', page.title);
            divsuggestion.setAttribute('data-uri', baseUri + page.uri);
            divsuggestion.setAttribute('data-context', context);
            divsuggestion.innerText = '» ' + page.title;
            divsuggestion.appendChild(divcontext);
            return divsuggestion.outerHTML;
        },
        /* onSelect callback fires when a search suggestion is chosen */
        onSelect: function(e, term, item) {
            location.href = item.getAttribute('data-uri');
            e.preventDefault();
        }
    });

    // JavaScript-autoComplete only registers the focus event when minChars is 0 which doesn't make sense, let's do it ourselves
    // https://github.com/Pixabay/JavaScript-autoComplete/blob/master/auto-complete.js#L191
    var selector = $('#search-by').get(0);
    $(selector).focus(selector.focusHandler);
});
